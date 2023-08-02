# TStat

Tstat provides information on Go test suites and functionality to make it easier to
query for information on code coverage or test cases.

For tests, it leverages the JSON output provides by `go test` when running with the `-json` flag. See [here](https://pkg.go.dev/cmd/go/internal/test) for more info.

For coverage, it leverages the cover profiles for statements and function coverage provided by the `cover` tool. See [cover](https://pkg.go.dev/cmd/cover) for more info.

## Usage

### Coverage
```go
	stats, err := tstat.Cover("testdata/prog/cover.out", tstat.WithRootModule("github.com/nickfiggins/tstat"))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("total coverage: %#v%%\n", stats.Percent)
	pkg := stats.Packages[0]
	fmt.Printf("package: %s coverage: %#v%%\n", pkg.Name, pkg.Percent)
	fileCov := pkg.Files[0]
	for _, fn := range fileCov.Functions {
		fmt.Printf("function: %s coverage: %v%%\n", fn.Name, fn.Percent)
	}
	// Output:
	// total coverage: 25%
	// package: github.com/nickfiggins/tstat/testdata/prog coverage: 25%
	// function: add coverage: 100%
	// function: isOdd coverage: 0%
```

### Tests

```go
	stats, err := tstat.Tests("testdata/bigtest.json")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(stats.Count(), stats.Failed(), stats.Duration().String())
	pkg, _ := stats.Package("github.com/nickfiggins/tstat")
	test, _ := pkg.Test("Test_CoverageStats")
	fmt.Println(test.Count(), test.Failed(), test.Skipped(), test.Package)

	sub, _ := test.Test("happy") // subtest Test_CoverageStats/happy
	if !sub.Failed() {
		fmt.Printf("%v passed\n", sub.FullName)
	}
	// Output: 50 false 473.097ms
	// 3 false false github.com/nickfiggins/tstat
	// Test_CoverageStats/happy passed
```