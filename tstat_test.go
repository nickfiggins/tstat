package tstat_test

import (
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/nickfiggins/tstat"
)

func TestRead(t *testing.T) {
	testDir := "testdata"
	cov, _ := os.Open(testDir + "/cover.out")
	f, _ := os.Open(testDir + "/func.out")
	type args struct {
		cov io.Reader
		fn  io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *tstat.Stats
		wantErr bool
	}{
		{
			name: "happy",
			args: args{cov: cov, fn: f},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tstat.Read(tt.args.cov, tt.args.fn)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() = %+v, want %v", got, tt.want)
			}
		})
	}
}
