package priority

import "testing"

func TestAdd(t *testing.T) {
	m := make(Map)
	m.Add("apple", "orange")
	m.Add("apple", "banana")
	m.Add("apple", "orange")
	l := m["apple"]
	if len(l) != 2 {
		t.Errorf("Priority: expecting two superiors, got %d", len(l))
	}
}

func TestComplete(t *testing.T) {
	m := make(Map)
	m.Add("apple", "orange")
	m.Add("orange", "banana")
	m.Add("orange", "grapes")
	m.Add("banana", "grapes")
	m.Add("banana", "grapefruit")
	m.Complete()
	l := m["apple"]
	if len(l) != 4 {
		t.Errorf("Priority: expecting two superiors, got %d", len(l))
	}
	l = m["orange"]
	if len(l) != 3 {
		t.Errorf("Priority: expecting two superiors, got %d", len(l))
	}
}

func TestList(t *testing.T) {
	m := make(Map)
	m.Add("apple", "orange")
	m.Add("orange", "banana")
	m.Add("orange", "grapes")
	m.Add("banana", "grapes")
	m.Add("banana", "grapefruit")
	m.Complete()
	list := m.List([]string{"apple", "grapes", "grapes", "banana", "banana", "apple"})
	if len(list) != 6 {
		t.Errorf("Priority: expecting six sets of indexes, got %d", len(list))
	}
	if len(list[0]) != 4 {
		t.Errorf("Priority: expecting four indexes for apple, got %v", len(list[0]))
	}
	if len(list[5]) != 4 {
		t.Errorf("Priority: expecting four indexes for apple, got %v", len(list[5]))
	}
}

func TestSubset(t *testing.T) {
	m := make(Map)
	m.Add("apple", "orange")
	m.Add("orange", "banana")
	m.Add("orange", "grapes")
	m.Add("banana", "grapes")
	m.Add("banana", "grapefruit")
	m.Complete()
	list := m.List([]string{"apple", "grapes", "grapes", "banana", "banana", "apple"})
	sub := list.Subset([]int{0, 3, 5})
	if len(sub) != 3 {
		t.Errorf("Priority: expecting 3 in the subsdet list, got %d", len(sub))
	}
	if len(sub[0]) != 4 {
		t.Errorf("Priority: expecting four indexes for apple, got %v", len(sub[0]))
	}
	if len(sub[2]) != 4 {
		t.Errorf("Priority: expecting four indexes for apple, got %v", len(sub[2]))
	}
}
