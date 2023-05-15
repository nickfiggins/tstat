package tstat

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/nickfiggins/tstat/internal/gofunc"
	"github.com/nickfiggins/tstat/internal/gotest"
	"golang.org/x/sync/errgroup"
	"golang.org/x/tools/cover"
)

type Parser struct {
	opts        Options
	testParser  func(io.Reader) ([]gotest.Event, error)
	coverParser func(io.Reader) ([]*cover.Profile, error)
	funcParser  func(io.Reader) (gofunc.Output, error)
}

func NewParser(opts ...ParseOpt) *Parser {
	var options Options
	for _, opt := range opts {
		opt(&options)
	}

	return &Parser{
		opts:        options,
		testParser:  gotest.ReadJSON,
		coverParser: cover.ParseProfilesFromReader,
		funcParser:  gofunc.Read,
	}
}

type Options struct {
	trimModule string
	out        io.Writer
}

type ParseOpt func(*Options)

// TrimModule will trim the given prefix (likely the name of a package or module)
// from all object names returned.
func TrimModule(name string) ParseOpt {
	return func(o *Options) {
		o.trimModule = name
	}
}

func WriteTo(w io.Writer) ParseOpt {
	return func(o *Options) {
		o.out = w
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

	fnOut, err := runFuncCover(profile)
	if err != nil {
		return Coverage{}, err
	}

	output, err := p.funcParser(bytes.NewBuffer(fnOut))
	if err != nil {
		return Coverage{}, err
	}

	cov := Coverage{Statement: parseProfiles(profiles, p.opts), Function: parseFuncProfile(output, p.opts)}
	if p.opts.out != nil {
		err = cov.writeTo(p.opts.out)
		if err != nil {
			return Coverage{}, err
		}
	}

	return cov, nil
}

func (p *Parser) CoverageStatsFromReaders(profile, funcProfile io.Reader) (Coverage, error) {
	var stmtStats *StatementStats
	var fnStats *FunctionStats
	group := errgroup.Group{}
	group.Go(func() error {
		profiles, err := p.coverParser(profile)
		if err != nil {
			return fmt.Errorf("couldn't parse coverage from reader: %w", err)
		}
		stmtStats = parseProfiles(profiles, p.opts)
		return nil
	})

	group.Go(func() error {
		output, err := p.funcParser(funcProfile)
		if err != nil {
			return err
		}
		fnStats = parseFuncProfile(output, p.opts)
		return nil
	})

	err := group.Wait()
	if err != nil {
		return Coverage{}, err
	}

	cov := Coverage{Function: fnStats, Statement: stmtStats}
	if p.opts.out != nil {
		err = cov.writeTo(p.opts.out)
		if err != nil {
			return Coverage{}, err
		}
	}

	return cov, nil
}

func (p *Parser) TestRunFromReader(jsonOutput io.Reader) (TestRun, error) {
	out, err := p.testParser(jsonOutput)
	if err != nil {
		return TestRun{}, err
	}

	return parseTestOutputs(out)
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

	return parseTestOutputs(out)
}

func runFuncCover(profile string) ([]byte, error) {
	goTool := filepath.Join(runtime.GOROOT(), "bin/go")
	cmd := exec.Command(goTool, "tool", "cover", fmt.Sprintf("-func=%v", profile))
	fnProfile, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("couldn't get function coverage: %w", handleExecError(err))
	}
	return fnProfile, nil
}

func handleExecError(err error) error {
	ee := &exec.ExitError{}
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		return fmt.Errorf("%w, stderr %v", err, string(ee.Stderr))
	}
	return err
}
