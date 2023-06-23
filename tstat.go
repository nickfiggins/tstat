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

// CoverageParser is a parser for coverage profiles that can be configured to read from files or io.Readers.
// If only a cover profile is provided, the corresponding function profile will be generated automatically.
// If a function profile is provided, it will be used instead of generating one - which is useful when parsing profiles
// that aren't part of the current project.
type CoverageParser struct {
	TrimModule   string
	WriteTo      io.Writer
	FuncProfile  io.Reader
	CoverProfile io.Reader

	coverParser func(io.Reader) ([]*cover.Profile, error)
	funcParser  func(io.Reader) (gofunc.Output, error)
}

// Cover creates a new CoverageParser with the provided options. If no options are provided, the parser will attempt to
// read the cover and func profiles from the TSTAT_COVER_PROFILE and TSTAT_FUNC_PROFILE environment variables.
// If those variables are not set, the parser must receive a cover profile via the WithCoverProfile option, or both
// profiles via the WithFiles option.
func Cover(coverProfile string, opts ...CoverOpt) (*CoverageParser, error) {
	f, err := os.Open(coverProfile)
	if err != nil {
		return nil, err
	}
	fnOut, err := runFuncCover(coverProfile)
	if err != nil {
		return nil, err
	}

	return newCoverageParser(f, bytes.NewBuffer(fnOut), opts...)
}

func CoverFromReaders(coverProfile io.Reader, fnProfile io.Reader, opts ...CoverOpt) (*CoverageParser, error) {
	return newCoverageParser(coverProfile, fnProfile, opts...)
}

func newCoverageParser(cov io.Reader, fn io.Reader, opts ...CoverOpt) (*CoverageParser, error) {
	parser := &CoverageParser{
		WriteTo:      nil,
		coverParser:  cover.ParseProfilesFromReader,
		funcParser:   gofunc.Read,
		FuncProfile:  fn,
		CoverProfile: cov,
	}

	for _, opt := range opts {
		if err := opt(parser); err != nil {
			return nil, err
		}
	}

	return parser, nil
}

// CoverOpt is a functional option for configuring a CoverageParser.
type CoverOpt func(*CoverageParser) error

// WithRootModule sets the root module to trim from the file names in the coverage profile.
func WithRootModule(module string) CoverOpt {
	return func(cp *CoverageParser) error {
		cp.TrimModule = filepath.Clean(module)
		return nil
	}
}

func (p *CoverageParser) fileName(full string) string {
	if p.TrimModule == "" {
		return full
	}
	return strings.TrimPrefix(full, p.TrimModule+"/")
}

// Stats parses the coverage and function profiles and returns a statistics based on the profiles read.
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

// TestsFromReader creates a new TestParser with the provided io.Reader.
func TestsFromReader(outJSON io.Reader) *TestParser {
	return &TestParser{
		jsonOut:    outJSON,
		testParser: gotest.ReadJSON,
	}
}

// Tests creates a new TestParser with the provided file.
func Tests(outFile string) (*TestParser, error) {
	f, err := os.Open(outFile)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %w", err)
	}
	return &TestParser{
		jsonOut:    f,
		testParser: gotest.ReadJSON,
	}, nil
}

// Stats parses the test output and returns a TestRun based on the output read.
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
