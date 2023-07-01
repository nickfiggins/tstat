package tstat_test

import (
	"errors"
	"strings"
	"testing"

	"encoding/json"

	"github.com/nickfiggins/tstat"
)

func Fuzz_TestRun_Stats(f *testing.F) {
	f.Add(`mode: set1234`)
	f.Fuzz(func(t *testing.T, testOut string) {
		_, err := tstat.TestsFromReader(strings.NewReader(testOut))
		var se *json.SyntaxError
		var de *json.UnmarshalTypeError
		if err != nil && !errors.As(err, &se) && !errors.As(err, &de) {
			t.Fatal(err)
		}
	})
}
