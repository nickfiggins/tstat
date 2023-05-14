package gotest

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func format(t *testing.T, ts string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestReadJSON(t *testing.T) {
	tests := []struct {
		name    string
		have    io.Reader
		want    []Output
		wantErr bool
	}{
		{
			name: "simple start, end",
			have: strings.NewReader(`
				{"Time":"2023-05-13T21:30:15.409912-04:00","Action":"start","Package":"github.com/nickfiggins/tstat/testdata/prog"}
				{"Time":"2023-05-13T21:30:15.59089-04:00","Action":"pass","Package":"github.com/nickfiggins/tstat/testdata/prog","Elapsed":0.181}
				`),
			want: []Output{
				{
					Time:    format(t, "2023-05-13T21:30:15.409912-04:00"),
					Action:  "start",
					Package: "github.com/nickfiggins/tstat/testdata/prog",
				},
				{
					Time:    format(t, "2023-05-13T21:30:15.59089-04:00"),
					Action:  "pass",
					Package: "github.com/nickfiggins/tstat/testdata/prog",
				},
			},
			wantErr: false,
		},
		{
			name: "simple pass and fail",
			have: strings.NewReader(`
				{"Time":"2023-05-13T21:30:15.587441-04:00","Action":"run","Package":"github.com/nickfiggins/tstat/testdata/prog","Test":"TestAdd"}
				{"Time":"2023-05-13T21:30:15.587512-04:00","Action":"output","Package":"github.com/nickfiggins/tstat/testdata/prog","Test":"TestAdd","Output":"=== RUN   TestAdd\n"}
				{"Time":"2023-05-13T21:30:15.587549-04:00","Action":"output","Package":"github.com/nickfiggins/tstat/testdata/prog","Test":"TestAdd","Output":"--- PASS: TestAdd (0.00s)\n"}
				{"Time":"2023-05-13T21:30:15.587555-04:00","Action":"pass","Package":"github.com/nickfiggins/tstat/testdata/prog","Test":"TestAdd","Elapsed":0}
				`),
			want: []Output{
				{
					Time:    format(t, "2023-05-13T21:30:15.587441-04:00"),
					Action:  "run",
					Package: "github.com/nickfiggins/tstat/testdata/prog",
					Test:    "TestAdd",
				},
				{
					Time:    format(t, "2023-05-13T21:30:15.587512-04:00"),
					Action:  "output",
					Package: "github.com/nickfiggins/tstat/testdata/prog",
					Output:  "=== RUN   TestAdd\n",
					Test:    "TestAdd",
				},
				{
					Time:    format(t, "2023-05-13T21:30:15.587549-04:00"),
					Action:  "output",
					Package: "github.com/nickfiggins/tstat/testdata/prog",
					Output:  "--- PASS: TestAdd (0.00s)\n",
					Test:    "TestAdd",
				},
				{
					Time:    format(t, "2023-05-13T21:30:15.587555-04:00"),
					Action:  "pass",
					Package: "github.com/nickfiggins/tstat/testdata/prog",
					Test:    "TestAdd",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadJSON(tt.have)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAction_String(t *testing.T) {
	tests := []struct {
		have Action
		want string
	}{
		{Pass, "pass"},
		{Fail, "fail"},
		{Out, "output"},
		{Skip, "skip"},
		{Start, "start"},
		{Action(-1), "invalid status"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.have.String(); got != tt.want {
				t.Errorf("Action.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
