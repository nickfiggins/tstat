package tstat_test

import (
	"fmt"

	"github.com/nickfiggins/tstat"
)

// TODO

func ExampleCover() {
	parser := tstat.NewParser(tstat.TrimModule("github.com/nickfiggins/tstat/"))

	stats, _ := parser.CoverageStats("cover.out")
	fmt.Printf("total coverage: %#v%%\n", stats.Function.CoverPct)
	fileCov, ok := (stats.Function.File("cover.go"))
	if ok {
		for _, fn := range fileCov {
			fmt.Printf("function: %s coverage: %v%%\n", fn.Name, fn.CoverPct)
		}
	}
	// Output:
	// 	total coverage: 72%
	// function: parseProfiles coverage: 100%
	// function: percent coverage: 66.7%
	// function: parseProfile coverage: 100%
	// function: addFunc coverage: 100%
	// function: Func coverage: 0%
	// function: ParseFuncProfile coverage: 0%
	// function: ParseFuncProfileFromReader coverage: 85.7%
	// function: File coverage: 0%
}

func ExampleTest() {
	parser := tstat.NewParser()

	stats, _ := parser.TestRun("test-run-102122.json")
	fmt.Println(stats)
	// Output: 2
}
