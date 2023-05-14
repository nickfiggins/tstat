# TStat

Tstat provides information on Go test suites and functionality to make it easier to
query for information on code coverage or test cases.

## Usage

### Coverage
```go
	parser := tstat.NewParser(tstat.TrimModule("github.com/nickfiggins/tstat/"))

	stats, err := parser.CoverageStats("testdata/prog/cover.out")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("total coverage: %#v%%\n", stats.Function.CoverPct)
	fileCov, ok := stats.Function.File("testdata/prog/prog.go")
	if !ok {
		...
	}
	
	for _, fn := range fileCov {
		fmt.Printf("function: %s coverage: %v%%\n", fn.Name, fn.CoverPct)
	}
	// Output:
	// total coverage: 25%
	// function: add coverage: 100%
	// function: isOdd coverage: 0%
```

### Tests

```go
	parser := tstat.NewParser()

	stats, err := parser.TestRun("testdata/prog/test.json")
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(stats.Count())
	// Output: 13

    fmt.Printf("%+v\n", stats.Tests[0])
    // Output: &{Subtests:[] Action:fail Name:TestHandleQuestion_Error SubName:TestHandleQuestion_Error Package:github.com/nickfiggins/  gotestsimulate/gogeo}
```