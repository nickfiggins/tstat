package tstat_test

import (
	"fmt"
	"log"

	"github.com/nickfiggins/tstat"
)

func ExampleCover() {
	stats, err := tstat.Cover("testdata/prog/cover.out", tstat.WithRootModule("github.com/nickfiggins/tstat"))
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("total coverage: %#v%%\n", stats.Percent)
	fileCov := stats.Packages[0].Files[0]
	for _, fn := range fileCov.Functions {
		fmt.Printf("function: %s coverage: %v%%\n", fn.Name, fn.Percent)
	}
	// Output:
	// total coverage: 25%
	// function: add coverage: 100%
	// function: isOdd coverage: 0%
}

func ExampleTests() {
	stats, err := tstat.Tests("testdata/bigtest.json")
	if err != nil {
		log.Fatalln(err)
	}
	// TODO: display more functionality here

	fmt.Println(stats.Count(), stats.Failed(), stats.Duration().String())
	pkg, _ := stats.Package("github.com/nickfiggins/tstat")
	test, _ := pkg.Test("Test_CoverageStats")
	fmt.Println(test.Count(), test.Failed(), test.Package)
	// Output: 50 false 473.097ms
	// 3 false github.com/nickfiggins/tstat
}
