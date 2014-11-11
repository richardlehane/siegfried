package pronom

import "testing"

func TestPlace(t *testing.T) {
	test := []string{"apple", "apple", "cherry", "banana", "banana", "banana"}
	if x, y := place(3, test); x != 1 || y != 3 {
		t.Errorf("Identifier: place error - expecting 1, 3 got %d, %d", x, y)
	}
	if x, y := place(4, test); x != 2 || y != 3 {
		t.Errorf("Identifier: place error - expecting 2, 3 got %d, %d", x, y)
	}
	if x, y := place(5, test); x != 3 || y != 3 {
		t.Errorf("Identifier: place error - expecting 3, 3 got %d, %d", x, y)
	}
	if x, y := place(0, test); x != 1 || y != 2 {
		t.Errorf("Identifier: place error - expecting 1, 2 got %d, %d", x, y)
	}
}
