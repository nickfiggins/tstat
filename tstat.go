package tstat

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/sync/errgroup"
)

type Parser struct {
	opts []ParseOpts
}

func NewParser(opts ...ParseOpts) *Parser {
	return &Parser{opts}
}

type Options struct {
	trimModule string
}

type ParseOpts func(*Options)

func TrimModule(name string) ParseOpts {
	return func(o *Options) {
		o.trimModule = name
	}
}

func (o Options) fileName(full string) string {
	if o.trimModule == "" {
		return strings.TrimPrefix(full, o.trimModule)
	}
	return full
}

func (p *Parser) CoverageStats(profile string) (*Coverage, error) {
	pf, err := os.Open(profile)
	if err != nil {
		return nil, fmt.Errorf("couldn't open cover profile: %w", err)
	}
	covStats, err := ParseCoverProfileFromReader(pf, p.opts...)
	if err != nil {
		return nil, err
	}

	goTool := filepath.Join(runtime.GOROOT(), "bin/go")
	cmd := exec.Command(goTool, "tool", "cover", fmt.Sprintf("-func=%v", profile))
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr
	fnProfile, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("couldn't get function coverage: %w, stderr %v", err, stderr.String())
	}

	fnStats, err := ParseFuncProfileFromReader(bytes.NewBuffer(fnProfile), p.opts...)
	if err != nil {
		return nil, err
	}

	return &Coverage{Statement: covStats, Function: fnStats}, nil
}

func (p *Parser) CoverageStatsFromReaders(profile, funcProfile io.Reader, opts ...ParseOpts) (*Coverage, error) {
	opts = append(opts, p.opts...)

	var stmtStats *StatementStats
	var fnStats *FunctionStats
	eg := errgroup.Group{}
	eg.Go(func() error {
		stmt, err := ParseCoverProfileFromReader(profile, opts...)
		if err != nil {
			return fmt.Errorf("couldn't parse cover profile: %w", err)
		}
		stmtStats = stmt
		return nil
	})

	eg.Go(func() error {
		fn, err := ParseFuncProfileFromReader(funcProfile, opts...)
		if err != nil {
			return fmt.Errorf("couldn't parse function profile: %w", err)
		}
		fnStats = fn
		return nil
	})

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	return &Coverage{Function: fnStats, Statement: stmtStats}, nil
}

func (p *Parser) TestStatsFromReader(jsonOutput io.Reader) (*TestStats, error) {
	return ParseTestOutput(jsonOutput)
}

func (p *Parser) TestStats(outputFile string) (*TestStats, error) {
	of, err := os.Open(outputFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't open test output file: %w", err)
	}

	defer of.Close()

	return ParseTestOutput(of)
}
