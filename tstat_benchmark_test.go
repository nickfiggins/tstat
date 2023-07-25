package tstat_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/nickfiggins/tstat"
)

var sink any

func BenchmarkTests(b *testing.B) {
	testJSON, err := os.ReadFile("testdata/go-cmp/go-cmp.json")
	if err != nil {
		b.Fatal("failed to read test data", err)
	}
	buf := bytes.NewBuffer(testJSON)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr, err := tstat.TestsFromReader(buf)
		if err != nil {
			b.Fatal("failed to parse test data", err)
		}
		sink = tr
	}
}
