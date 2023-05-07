package tstat_test

import (
	"fmt"
	"testing"

	"github.com/nickfiggins/tstat"
)

func TestXxx(t *testing.T) {
	cov, _ := tstat.ParseFuncProfile("./testdata/func.out")
	fmt.Println("data", cov.File("db.go"))
}
