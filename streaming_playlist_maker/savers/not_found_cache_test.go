package savers

import (
	"testing"
)

func TestNotFound_empty(t *testing.T) {
	n := newCache()
	s := n.IsNotFound("some song")
	if s != nil {
		t.Errorf("NotFound cache should be empty after start")
	}
}

func TestNotFound(t *testing.T) {
	n := newCache()
	n.AddNotFound("some song", &Status{FoundTitle: "Some title"})
	s := n.IsNotFound("some song")
	if s.FoundTitle != "Some title" {
		t.Errorf("NotFound cache should contain 'some song'")
	}
}
