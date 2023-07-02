package gotest

import "testing"

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
