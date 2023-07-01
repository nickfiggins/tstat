# TStat

Tstat provides information on Go test suites and functionality to make it easier to
query for information on code coverage or test cases.

## Usage

### Coverage
```go
	stats, err := tstat.Cover("testdata/prog/cover.out", tstat.WithRootModule("github.com/nickfiggins/tstat"))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("total coverage: %#v%%\n", stats.Function.CoverPct)
	fileCov, ok := stats.Function.File("testdata/prog/prog.go")
	if ok {
		for _, fn := range fileCov {
			fmt.Printf("function: %s coverage: %v%%\n", fn.Name, fn.CoverPct)
		}
	}
	// Output:
	// total coverage: 25%
	// function: add coverage: 100%
	// function: isOdd coverage: 0%
```

### Tests

```go
	stats, err := tstat.Tests("testdata/bigtest.json")
	if err != nil {
		log.Fatalln(err)
	}
	// TODO: display more functionality here

	fmt.Println(stats.Count(), stats.Failed(), stats.Duration().String())
	pkg, _ := stats.Package("github.com/nickfiggins/tstat")
	test := pkg.Test("Test_CoverageStats")
	fmt.Println(test.Count(), test.Failed())
	// Output: 50 false 473.097ms
	// 3 false
```