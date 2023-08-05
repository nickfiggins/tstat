package tstat

import (
	"errors"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/nickfiggins/tstat/internal/gocover"
	"github.com/nickfiggins/tstat/internal/gofunc"
	"github.com/nickfiggins/tstat/internal/gotest"
	"github.com/stretchr/testify/assert"
)

func Test_TestParser_Stats(t *testing.T) {
	fileName := t.TempDir() + "/" + "test.json"
	if _, err := os.Create(fileName); err != nil {
		t.Fatal(err)
	}
	testFile, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name    string
		parser  TestParser
		want    []PackageRun
		wantErr bool
	}{
		{
			name: "happy, nested subtests",
			parser: TestParser{
				testParser: func(r io.Reader) ([]*gotest.PackageEvents, error) {
					return []*gotest.PackageEvents{
						{
							Package: "pkg",
							Start:   nil, End: nil,
							Events: []gotest.Event{
								{Time: time.Now(), Action: gotest.Pass, Package: "pkg", Test: "Test1"},
							},
						},
						{
							Package: "pkg2",
							Start:   nil, End: nil,
							Events: []gotest.Event{
								{Time: time.Now(), Action: gotest.Pass, Package: "pkg2", Test: "Test2"},
							},
						},
					}, nil
				},
				converter: func(pkg *gotest.PackageEvents) (PackageRun, error) {
					return PackageRun{pkgName: pkg.Package}, nil
				},
			},
			want: []PackageRun{
				{pkgName: "pkg"},
				{pkgName: "pkg2"},
			},
		},
		{
			name: "error parsing tests",
			parser: TestParser{
				testParser: func(r io.Reader) ([]*gotest.PackageEvents, error) {
					return nil, errors.New("error parsing")
				},
			},
			want:    []PackageRun{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := tt.parser
			got, err := cp.Stats(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Stats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Slice(tt.want, func(i, j int) bool {
				return tt.want[i].pkgName > tt.want[j].pkgName
			})
			sort.Slice(got.pkgs, func(i, j int) bool {
				return got.pkgs[i].pkgName > got.pkgs[j].pkgName
			})
			for i, p := range tt.want {
				assert.ElementsMatch(t, p.Tests, got.pkgs[i].Tests)
			}
		})
	}
}

func Test_CoverageParser_Stats(t *testing.T) {
	fileName := t.TempDir() + "/" + "test.json"
	if _, err := os.Create(fileName); err != nil {
		t.Fatal(err)
	}
	testFile, err := os.Open(fileName)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		funcProfile io.Reader
		parser      CoverageParser
		want        Coverage
		wantErr     bool
	}{
		{
			name: "happy cov",
			parser: CoverageParser{
				coverParser: func(r io.Reader) ([]*gocover.PackageStatements, error) {
					return []*gocover.PackageStatements{
						{
							Package: "",
							Files: map[string]*gocover.FileStatements{
								"prog.go": {Percent: 20, Stmts: 5, CoveredStmts: 1},
							},
							Percent:      20,
							Stmts:        5,
							CoveredStmts: 1,
						},
					}, nil
				},
				funcParser: func(r io.Reader) ([]*gofunc.PackageFunctions, error) { return make([]*gofunc.PackageFunctions, 0), nil },
			},
			funcProfile: strings.NewReader(""),
			want: Coverage{
				Percent: 20,
				Packages: []*PackageCoverage{
					{
						Name:    "",
						Percent: 20,
						Files: []*FileCoverage{
							{
								Name:         "prog.go",
								Functions:    []FunctionCoverage{},
								Percent:      20,
								Stmts:        5,
								CoveredStmts: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "happy cov, trim module",
			parser: CoverageParser{
				coverParser: func(r io.Reader) ([]*gocover.PackageStatements, error) {
					return []*gocover.PackageStatements{
						{
							Package: "github.com/mod",
							Files: map[string]*gocover.FileStatements{
								"github.com/mod/prog.go": {Percent: 20, Stmts: 5, CoveredStmts: 1},
							},
							Percent:      10,
							Stmts:        5,
							CoveredStmts: 1,
						},
					}, nil
				},
				funcParser: func(r io.Reader) ([]*gofunc.PackageFunctions, error) {
					return []*gofunc.PackageFunctions{
						{
							Package: "github.com/mod",
							Files: map[string]*gofunc.FileFunctions{"github.com/mod/prog.go": {
								File: "github.com/mod/prog.go",
								Functions: []gofunc.Function{
									{Package: "github.com/mod", Function: "main", Percent: 10, Line: 1, File: "github.com/mod/prog.go"},
								},
							}},
						},
					}, nil

				},
			},
			funcProfile: strings.NewReader(""),
			want: Coverage{
				Percent: 20,
				Packages: []*PackageCoverage{
					{
						Name:    "github.com/mod",
						Percent: 20,
						Files: []*FileCoverage{
							{
								Name: "github.com/mod/prog.go",
								Functions: []FunctionCoverage{
									{Name: "main", Percent: 10, Line: 1, File: "github.com/mod/prog.go", Internal: true},
								},
								Percent:      20,
								Stmts:        5,
								CoveredStmts: 1,
							},
						},
					},
				},
			},
		},
		{
			name: "error reading coverage",
			parser: CoverageParser{
				coverParser: func(r io.Reader) ([]*gocover.PackageStatements, error) {
					return nil, errors.New("error parsing")
				},
				funcParser: func(r io.Reader) ([]*gofunc.PackageFunctions, error) {
					return make([]*gofunc.PackageFunctions, 0), nil
				},
			},
			funcProfile: strings.NewReader(""),
			want:        Coverage{},
			wantErr:     true,
		},
		{
			name: "error reading func profile",
			parser: CoverageParser{
				coverParser: func(r io.Reader) ([]*gocover.PackageStatements, error) {
					return []*gocover.PackageStatements{}, nil
				},
				funcParser: func(r io.Reader) ([]*gofunc.PackageFunctions, error) {
					return nil, errors.New("error parsing")
				},
			},
			funcProfile: &errReader{},
			want:        Coverage{},
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parser.Stats(testFile, tt.funcProfile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

type errReader struct{}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("err reading")
}
