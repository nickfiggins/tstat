package gotest

import (
	"testing"
	"time"
)

func Test_ToAction(t *testing.T) {
	tests := []struct {
		have string
		want Action
	}{
		{"PASS", Pass}, {"FAIL", Fail}, {"fAiL", Fail},
		{"output", Out}, {"skip", Skip}, {"start", Start},
		{"dfioroiriooi", Undefined}, {"undefined", Undefined},
	}
	for _, tt := range tests {
		if got := ToAction(tt.have); got != tt.want {
			t.Errorf("ToAction(%v) = %v, want %v", tt.have, got, tt.want)
		}
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
		{Action(-1), "undefined"},
	}
	for _, tt := range tests {
		if got := tt.have.String(); got != tt.want {
			t.Errorf("Action(%v).String() = %v, want %v", tt.have, got, tt.want)
		}
	}
}

func TestAction_IsFinal(t *testing.T) {
	tests := []struct {
		have Action
		want bool
	}{
		{Pass, true},
		{Fail, true},
		{Out, false},
		{Skip, true},
		{Start, false},
		{Action(-1), false},
	}
	for _, tt := range tests {
		if got := tt.have.IsFinal(); got != tt.want {
			t.Errorf("Action(%v).IsFinal() = %v, want %v", tt.have, got, tt.want)
		}
	}
}

func TestEvent_Seed(t *testing.T) {
	type fields struct {
		Time    time.Time
		Action  Action
		Output  string
		Test    string
		Package string
		Elapsed float64
	}
	tests := []struct {
		name   string
		fields fields
		want   int64
	}{
		{
			name: "no seed",
			fields: fields{
				Time:    time.Now(),
				Action:  Start,
				Output:  "some output",
				Test:    "some test",
				Package: "some package",
				Elapsed: 0.0,
			},
			want: 0,
		},
		{
			name: "no seed",
			fields: fields{
				Time:    time.Now(),
				Action:  Start,
				Output:  "-test.shuffle 123",
				Test:    "some test",
				Package: "some package",
				Elapsed: 0.0,
			},
			want: 123,
		},
		{
			name: "seed isn't an int",
			fields: fields{
				Time:    time.Now(),
				Action:  Start,
				Output:  "-test.shuffle oiddoifeoi",
				Test:    "some test",
				Package: "some package",
				Elapsed: 0.0,
			},
			want: 0,
		},
		{
			name: "no seed",
			fields: fields{
				Time:    time.Now(),
				Action:  Start,
				Output:  "-test.shuffle 1688261989310323000\n",
				Test:    "some test",
				Package: "some package",
				Elapsed: 0.0,
			},
			want: 1688261989310323000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Event{
				Time:    tt.fields.Time,
				Action:  tt.fields.Action,
				Output:  tt.fields.Output,
				Test:    tt.fields.Test,
				Package: tt.fields.Package,
				Elapsed: tt.fields.Elapsed,
			}
			got, gotOK := e.Seed()
			if got != tt.want {
				t.Errorf("Event.Seed() got = %v, want %v", got, tt.want)
			}

			if got == 0 && gotOK {
				t.Error("got 0, gotOK true")
			}
		})
	}
}
