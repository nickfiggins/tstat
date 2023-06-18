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
	"golang.org/x/tools/cover"
)

type CoverageParser struct {
	TrimModule   string
	WriteTo      io.Writer
	FuncProfile  io.Reader
	CoverProfile io.Reader

	coverParser func(io.Reader) ([]*cover.Profile, error)
	funcParser  func(io.Reader) (gofunc.Output, error)
}

func WithRootModule(module string) CoverOpt {
	return func(cp *CoverageParser) error {
		cp.TrimModule = filepath.Clean(module)
		return nil
	}
}

func WithFiles(cover, fn io.Reader) CoverOpt {
	return func(cp *CoverageParser) error {
		cp.CoverProfile = cover
		cp.FuncProfile = fn
		return nil
	}
}

func WithCoverProfile(cover string) CoverOpt {
	return func(cp *CoverageParser) error {
		f, err := os.Open(cover)
		if err != nil {
			return err
		}
		cp.CoverProfile = f
		fnOut, err := runFuncCover(cover)
		if err != nil {
			return err
		}
		cp.FuncProfile = bytes.NewBuffer(fnOut)
		return nil
	}
}

func Cover(opts ...CoverOpt) (*CoverageParser, error) {
	parser := &CoverageParser{
		WriteTo:     nil,
		coverParser: cover.ParseProfilesFromReader,
		funcParser:  gofunc.Read,
	}

	for _, opt := range opts {
		if err := opt(parser); err != nil {
			return nil, err
		}
	}

	return parser, nil
}

type CoverOpt func(*CoverageParser) error

func (p *CoverageParser) fileName(full string) string {
	if p.TrimModule == "" {
		return full
	}
	return strings.TrimPrefix(full, p.TrimModule+"/")
}

func (p *CoverageParser) Stats() (Coverage, error) {
	profiles, err := p.coverParser(p.CoverProfile)
	if err != nil {
		return Coverage{}, err
	}

	output, err := p.funcParser(p.FuncProfile)
	if err != nil {
		return Coverage{}, err
	}

	cov := Coverage{Statement: parseProfiles(profiles, p.fileName), Function: parseFuncProfile(output, p.fileName)}
	if p.WriteTo != nil {
		err = cov.writeTo(p.WriteTo)
		if err != nil {
			return Coverage{}, err
		}
	}

	return cov, nil
}

type TestParser struct {
	jsonOut    io.Reader
	testParser func(io.Reader) ([]gotest.Event, error)
}

func Tests(outJSON io.Reader) *TestParser {
	return &TestParser{
		jsonOut:    outJSON,
		testParser: gotest.ReadJSON,
	}
}

func TestsFromFile(outFile string) (*TestParser, error) {
	f, err := os.Open(outFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %w", err)
	}
	return &TestParser{
		jsonOut:    f,
		testParser: gotest.ReadJSON,
	}, nil
}

func (p *TestParser) Stats() (TestRun, error) {
	out, err := p.testParser(p.jsonOut)
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
	var ee *exec.ExitError
	if errors.As(err, &ee) && len(ee.Stderr) > 0 {
		return fmt.Errorf("%w, stderr %v", err, string(ee.Stderr))
	}
	return err
}
