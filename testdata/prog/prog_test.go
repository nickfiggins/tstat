package prog

import "testing"

func TestAdd(t *testing.T) {
	got := add(1, 2)
	if got != 3 {
		t.Fail()
	}
}
