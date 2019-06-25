package main

import (
	"testing"
)

func TestInit(t *testing.T) {
	s := &statistics{}
	err := s.Init("name")
	if err != nil {
		t.Errorf("Init: got: %v want: nil", err)
	}
	// Second init with the same name
	err = s.Init("name")
	if err == nil {
		t.Errorf("Init: got: nil want: some_error")
	}
}

func TestAdding(t *testing.T) {
	s := &statistics{}
	s.Init("name")
	s.Added("name", "")
	if s.m["name"].added != 1 {
		t.Errorf("TestAdding: got: %v want: 1", s.m["name"].added)
	}

	s.Added("name", "")
	if s.m["name"].added != 2 {
		t.Errorf("TestAdding: got: %v want: 2", s.m["name"].added)
	}
}

func TestNotFound(t *testing.T) {
	s := &statistics{}
	s.Init("name")
	s.NotFound("name", "")
	if s.m["name"].notFound != 1 {
		t.Errorf("TestNf: got: %v, want: 1", s.m["name"].notFound)
	}

	s.NotFound("name", "")
	if s.m["name"].notFound != 2 {
		t.Errorf("TestNf: got: %v, want: 2", s.m["name"].notFound)
	}
}

func TestExists(t *testing.T) {
	s := &statistics{}
	s.Init("name")
	s.Exists("name", "")
	if s.m["name"].exists != 1 {
		t.Errorf("TestEx: got: %v want: 1", s.m["name"].exists)
	}

	s.Exists("name", "")
	if s.m["name"].exists != 2 {
		t.Errorf("TestEx: got: %v want: 2", s.m["name"].exists)
	}
}

func TestError(t *testing.T) {
	s := &statistics{}
	s.Init("name")
	s.Error("name", "", nil)
	if s.m["name"].errors != 1 {
		t.Errorf("TestErr: got: %v want: 1", s.m["name"].errors)
	}

	s.Error("name", "", nil)
	if s.m["name"].errors != 2 {
		t.Errorf("TestErr: got: %v want: 2", s.m["name"].errors)
	}
}
