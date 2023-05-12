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

	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/cover"
)

type Parser struct {
	opts        []ParseOpts
	testParser  func(io.Reader) ([]gotest.Output, error)
	coverParser func(io.Reader) ([]*cover.Profile, error)
}

func NewParser(opts ...ParseOpts) *Parser {
	return &Parser{
		opts:        opts,
		testParser:  gotest.ReadJSON,
		coverParser: cover.ParseProfilesFromReader,
	}
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
	if o.trimModule != "" {
		return strings.TrimPrefix(full, o.trimModule)
	}
	return full
}

func (p *Parser) CoverageStats(profile string) (Coverage, error) {
	pf, err := os.Open(profile)
	if err != nil {
		return Coverage{}, fmt.Errorf("couldn't open cover profile: %w", err)
	}
	profiles, err := p.coverParser(pf)
	if err != nil {
		return Coverage{}, err
	}
	covStats := parseProfiles(profiles, p.opts...)

	fnOut, err := p.runFuncCover(profile)
	if err != nil {
		return Coverage{}, err
	}

	fnStats, err := ParseFuncProfileFromReader(bytes.NewBuffer(fnOut), p.opts...)
	if err != nil {
		return Coverage{}, err
	}

	return Coverage{Statement: &covStats, Function: &fnStats}, nil
}

func (p *Parser) runFuncCover(profile string) ([]byte, error) {
	goTool := filepath.Join(runtime.GOROOT(), "bin/go")
	cmd := exec.Command(goTool, "tool", "cover", fmt.Sprintf("-func=%v", profile))
	stderr := bytes.NewBuffer([]byte{})
	cmd.Stderr = stderr
	fnProfile, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("couldn't get function coverage: %w, stderr %v", err, stderr.String())
	}
	return fnProfile, nil
}

func (p *Parser) CoverageStatsFromReaders(profile, funcProfile io.Reader, opts ...ParseOpts) (Coverage, error) {
	opts = append(opts, p.opts...)

	var stmtStats StatementStats
	var fnStats FunctionStats
	group := errgroup.Group{}
	group.Go(func() error {
		profiles, err := p.coverParser(profile)
		if err != nil {
			return fmt.Errorf("couldn't parse coverage from reader: %w", err)
		}
		stmtStats = parseProfiles(profiles, opts...)
		return nil
	})

	group.Go(func() error {
		fn, err := ParseFuncProfileFromReader(funcProfile, opts...)
		if err != nil {
			return fmt.Errorf("couldn't parse function profile: %w", err)
		}
		fnStats = fn
		return nil
	})

	err := group.Wait()
	if err != nil {
		return Coverage{}, err
	}

	return Coverage{Function: &fnStats, Statement: &stmtStats}, nil
}

func (p *Parser) TestRunFromReader(jsonOutput io.Reader) (TestRun, error) {
	out, err := p.testParser(jsonOutput)
	if err != nil {
		return TestRun{}, err
	}

	return parseTestOutput(out)
}

func (p *Parser) TestRun(outputFile string) (TestRun, error) {
	of, err := os.Open(outputFile)
	if err != nil {
		return TestRun{}, fmt.Errorf("couldn't open test output file: %w", err)
	}
	defer of.Close()
	out, err := p.testParser(of)
	if err != nil {
		return TestRun{}, err
	}

	return parseTestOutput(out)
}
