package mathutil

import "testing"

func TestPercent(t *testing.T) {
	type args struct {
		num int64
		den int64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{name: "simple ints", args: args{num: 100, den: 1000}, want: 10},
		{name: "two-fifths", args: args{num: 2, den: 5}, want: 40},
		{name: "over 100", args: args{num: 11, den: 10}, want: 110},
		{name: "zero denom", args: args{num: 10, den: 0}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Percent(tt.args.num, tt.args.den); got != tt.want {
				t.Errorf("Percent() = %v, want %v", got, tt.want)
			}
		})
	}
}
