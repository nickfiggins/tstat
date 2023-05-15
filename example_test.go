package tstat_test

import (
	"fmt"
	"log"

	"github.com/nickfiggins/tstat"
)

func ExampleParser_CoverageStats() {
	parser := tstat.NewParser(tstat.TrimModule("github.com/nickfiggins/tstat/"))

	stats, err := parser.CoverageStats("testdata/prog/cover.out")
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
}

func ExampleParser_TestRun() {
	parser := tstat.NewParser()

	stats, err := parser.TestRun("testdata/prog/test.json")
	if err != nil {
		log.Fatalln(err)
	}
	// TODO: display more functionality here

	fmt.Println(stats.Count())
	// Output: 1
}
