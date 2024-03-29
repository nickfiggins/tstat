package gotest

import (
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadByPackage(t *testing.T) {
	tests := []struct {
		name    string
		have    io.Reader
		want    []*PackageEvents
		wantErr bool
	}{
		{
			name: "simple start, end",
			have: strings.NewReader(`
				{"Time":"2023-05-13T21:30:15.409912-04:00","Action":"start","Package":"github.com/nickfiggins/tstat/testdata/prog"}
				{"Time":"2023-05-13T21:30:15.59089-04:00","Action":"pass","Package":"github.com/nickfiggins/tstat/testdata/prog","Elapsed":0.181}
				`),
			want: []*PackageEvents{
				{
					Package: "github.com/nickfiggins/tstat/testdata/prog",
					Start:   &Event{Time: format(t, "2023-05-13T21:30:15.409912-04:00"), Action: Start, Package: "github.com/nickfiggins/tstat/testdata/prog"},
					End:     &Event{Time: format(t, "2023-05-13T21:30:15.59089-04:00"), Action: Pass, Package: "github.com/nickfiggins/tstat/testdata/prog", Elapsed: 0.181},
					Events: []Event{
						{Time: format(t, "2023-05-13T21:30:15.409912-04:00"), Action: Start, Package: "github.com/nickfiggins/tstat/testdata/prog"},
						{Time: format(t, "2023-05-13T21:30:15.59089-04:00"), Action: Pass, Package: "github.com/nickfiggins/tstat/testdata/prog", Elapsed: 0.181},
					},
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
			want: []*PackageEvents{
				{Package: "github.com/nickfiggins/tstat/testdata/prog", Events: []Event{
					{Time: format(t, "2023-05-13T21:30:15.587441-04:00"), Action: Run, Package: "github.com/nickfiggins/tstat/testdata/prog", Test: "TestAdd"},
					{Time: format(t, "2023-05-13T21:30:15.587512-04:00"), Action: Out, Package: "github.com/nickfiggins/tstat/testdata/prog", Test: "TestAdd", Output: "=== RUN   TestAdd\n"},
					{Time: format(t, "2023-05-13T21:30:15.587549-04:00"), Action: Out, Package: "github.com/nickfiggins/tstat/testdata/prog", Test: "TestAdd", Output: "--- PASS: TestAdd (0.00s)\n"},
					{Time: format(t, "2023-05-13T21:30:15.587555-04:00"), Action: Pass, Package: "github.com/nickfiggins/tstat/testdata/prog", Test: "TestAdd"},
				}},
			},
		},
		{
			name:    "error reading",
			have:    &errReader{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid json error",
			have:    strings.NewReader(`{"bad": "json}`),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadByPackage(tt.have)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadByPackage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

type errReader struct{}

func (e errReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func format(t *testing.T, ts string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

func TestByPackage(t *testing.T) {
	prog2Start := Event{
		Time:    format(t, "2023-05-13T21:30:15.587441-04:00"),
		Action:  Start,
		Package: "github.com/nickfiggins/tstat/testdata/prog2",
		Test:    "",
	}
	type args struct {
		events []Event
	}
	tests := []struct {
		name string
		args args
		want []*PackageEvents
	}{
		{
			name: "happy, single package",
			args: args{
				events: mockEvents(t),
			},
			want: []*PackageEvents{
				{
					Package: "github.com/nickfiggins/tstat/testdata/prog",
					Events:  mockEvents(t),
					Seed:    123,
					Start: &Event{
						Time:    format(t, "2023-05-13T21:30:15.587441-04:00"),
						Action:  Start,
						Package: "github.com/nickfiggins/tstat/testdata/prog",
						Test:    "",
					},
					End: nil,
				},
			},
		},
		{
			name: "happy, multi package",
			args: args{
				events: append(mockEvents(t), prog2Start),
			},
			want: []*PackageEvents{
				{
					Package: "github.com/nickfiggins/tstat/testdata/prog",
					Events:  mockEvents(t),
					Start: &Event{
						Time:    format(t, "2023-05-13T21:30:15.587441-04:00"),
						Action:  Start,
						Package: "github.com/nickfiggins/tstat/testdata/prog",
						Test:    "",
					},
					Seed: 123,
					End:  nil,
				},
				{
					Package: "github.com/nickfiggins/tstat/testdata/prog2",
					Events:  []Event{prog2Start},
					Start:   &prog2Start,
					End:     nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ByPackage(tt.args.events)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func mockEvents(t *testing.T) []Event {
	return []Event{
		{
			Time:    format(t, "2023-05-13T21:30:15.587441-04:00"),
			Action:  Start,
			Package: "github.com/nickfiggins/tstat/testdata/prog",
			Test:    "",
		},
		{
			Time:    format(t, "2023-05-13T21:30:15.587441-04:00"),
			Action:  Out,
			Package: "github.com/nickfiggins/tstat/testdata/prog",
			Output:  "-test.shuffle 123",
		},
		{
			Time:    format(t, "2023-05-13T21:30:15.587441-04:00"),
			Action:  Run,
			Package: "github.com/nickfiggins/tstat/testdata/prog",
			Test:    "TestAdd",
		},
		{
			Time:    format(t, "2023-05-13T21:30:15.587512-04:00"),
			Action:  Out,
			Package: "github.com/nickfiggins/tstat/testdata/prog",
			Output:  "=== RUN   TestAdd\n",
			Test:    "TestAdd",
		},
		{
			Time:    format(t, "2023-05-13T21:30:15.587549-04:00"),
			Action:  Out,
			Package: "github.com/nickfiggins/tstat/testdata/prog",
			Output:  "--- PASS: TestAdd (0.00s)\n",
			Test:    "TestAdd",
		},
		{
			Time:    format(t, "2023-05-13T21:30:15.587555-04:00"),
			Action:  Pass,
			Package: "github.com/nickfiggins/tstat/testdata/prog",
			Test:    "TestAdd",
		},
	}
}
