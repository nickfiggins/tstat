package tstat

import (
	"errors"
	"io"
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
	firstPkg := PackageRun{
		pkgName: "pkg",
		start:   zeroPlus(2),
		end:     zeroPlus(3),
		Tests: []*Test{
			{Subtests: []*Test{}, actions: []gotest.Action{gotest.Run, gotest.Pass}, Name: "Test", Package: "pkg", FullName: "Test"},
		},
	}
	secondPkg := PackageRun{
		pkgName: "pkg2",
		start:   zeroPlus(3),
		end:     zeroPlus(4),
		Tests: []*Test{
			{Subtests: []*Test{}, actions: []gotest.Action{gotest.Run, gotest.Pass}, Name: "Test", Package: "pkg2", FullName: "Test"},
		},
	}
	thirdPkg := PackageRun{
		pkgName: "pkg3",
		start:   zeroPlus(1),
		end:     zeroPlus(3),
		Tests:   []*Test{},
	}
	tests := []struct {
		name    string
		parser  TestParser
		want    TestRun
		wantErr bool
	}{
		{
			name: "happy",
			parser: TestParser{
				testParser: func(r io.Reader) ([]*gotest.PackageEvents, error) {
					return []*gotest.PackageEvents{{Package: "pkg"}, {Package: "pkg2"}}, nil
				},
				converter: func(pkg *gotest.PackageEvents) (PackageRun, error) {
					if pkg.Package == "pkg" {
						return firstPkg, nil
					}
					return secondPkg, nil
				},
			},
			want: TestRun{
				start: zeroPlus(2), // from first package
				end:   zeroPlus(4), // from second package
				pkgs:  []PackageRun{firstPkg, secondPkg},
			},
		},
		{
			name: "happy, last package started first",
			parser: TestParser{
				testParser: func(r io.Reader) ([]*gotest.PackageEvents, error) {
					return []*gotest.PackageEvents{{Package: "pkg"}, {Package: "pkg2"}, {Package: "pkg3"}}, nil
				},
				converter: func(pkg *gotest.PackageEvents) (PackageRun, error) {
					if pkg.Package == "pkg" {
						return firstPkg, nil
					}
					if pkg.Package == "pkg2" {
						return secondPkg, nil
					}
					return thirdPkg, nil
				},
			},
			want: TestRun{
				start: zeroPlus(1), // from third package
				end:   zeroPlus(4), // from second package
				pkgs:  []PackageRun{firstPkg, secondPkg, thirdPkg},
			},
		},
		{
			name: "error parsing tests",
			parser: TestParser{
				testParser: func(r io.Reader) ([]*gotest.PackageEvents, error) {
					return nil, errors.New("error parsing")
				},
			},
			want:    TestRun{},
			wantErr: true,
		},
		{
			name: "error converting",
			parser: TestParser{
				testParser: func(r io.Reader) ([]*gotest.PackageEvents, error) {
					return []*gotest.PackageEvents{{Package: "pkg"}}, nil
				},
				converter: func(pkg *gotest.PackageEvents) (PackageRun, error) {
					return PackageRun{}, errors.New("error converting")
				},
			},
			want:    TestRun{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cp := tt.parser
			got, err := cp.Stats(strings.NewReader(""))
			if (err != nil) != tt.wantErr {
				t.Errorf("Stats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			sort.Slice(tt.want.pkgs, func(i, j int) bool {
				return tt.want.pkgs[i].pkgName > tt.want.pkgs[j].pkgName
			})
			sort.Slice(got.pkgs, func(i, j int) bool {
				return got.pkgs[i].pkgName > got.pkgs[j].pkgName
			})
			assert.Equal(t, tt.want, got)
		})
	}
}

// zeroPlus returns a zero time with the given number of hours added
func zeroPlus(num int) time.Time {
	return time.Time{}.Add(time.Duration(num) * time.Hour)
}

func Test_CoverageParser_Stats(t *testing.T) {
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
			got, err := tt.parser.Stats(strings.NewReader(""), tt.funcProfile)
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
