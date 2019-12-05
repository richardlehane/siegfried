package patterns

import "testing"

func TestOverlap(t *testing.T) {
	res := overlap([]byte{'p', 'd', 'f', 'a'}, []byte{'d', 'f', 'a', 'b'})
	if res != 1 {
		t.Errorf("FAIL: expect 1, got %d\n", res)
	}
	res = overlap([]byte{'p', 'd', 'f', 'a'}, []byte{'f', 'a', 'b'})
	if res != 2 {
		t.Errorf("FAIL: expect 2, got %d\n", res)
	}
	res = overlap([]byte{'p', 'd', 'f', 'a'}, []byte{'a', 'b'})
	if res != 3 {
		t.Errorf("FAIL: expect 3, got %d\n", res)
	}
	res = overlap([]byte{'p', 'd', 'f', 'a'}, []byte{'b'})
	if res != 4 {
		t.Errorf("FAIL: expect 4, got %d\n", res)
	}
}

func TestOverlapR(t *testing.T) {
	res := overlapR([]byte{'e', 'o', 'f', 'a'}, []byte{'d', 'e', 'o', 'f'})
	if res != 1 {
		t.Errorf("FAIL: expect 1, got %d\n", res)
	}
	res = overlapR([]byte{'e', 'o', 'f', 'a'}, []byte{'d', 'e', 'o'})
	if res != 2 {
		t.Errorf("FAIL: expect 2, got %d\n", res)
	}
	res = overlapR([]byte{'e', 'o', 'f', 'a'}, []byte{'d', 'e'})
	if res != 3 {
		t.Errorf("FAIL: expect 3, got %d\n", res)
	}
	res = overlapR([]byte{'e', 'o', 'f', 'a'}, []byte{'a'})
	if res != 4 {
		t.Errorf("FAIL: expect 4, got %d\n", res)
	}
}
