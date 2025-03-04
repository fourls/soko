package crud_test

import (
	"testing"

	"github.com/fourls/soko/internal/crud"
)

func TestCrudReadAbsent(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	value, ok := crud.Read(1)
	if ok || value != "" {
		t.Fatalf("got: %v, %v when querying key, expected: %v, %v", value, ok, "", false)
	}
}

func TestCrudCreate(t *testing.T) {
	cases := []struct {
		key      int
		expected string
	}{
		{1, "foo"},
		{2, "bar"},
		{3, "  "},
		{-1, ""},
	}

	crud := crud.New[int, string]()
	defer crud.Close()

	for _, c := range cases {
		if !crud.Create(c.key, c.expected) {
			t.Fatalf("got: false when calling Create(%v, %v), expected: true", c.key, c.expected)
		}
	}

	for _, c := range cases {
		actual, ok := crud.Read(c.key)
		if !ok || actual != c.expected {
			t.Fatalf("got: %v, %v for key %d, expected: %v, true", actual, ok, c.key, c.expected)
		}
	}
}

func TestCrudCreateExisting(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	if !crud.Create(1, "foo") {
		t.Fatalf("got: false when calling first Create, expected: true")
	}
	if crud.Create(1, "bar") {
		t.Fatalf("got: true when calling second Create, expected: false")
	}

	value, ok := crud.Read(1)
	if !ok || value != "foo" {
		t.Fatalf("got: %v, %v when querying key, expected: %v, true", value, ok, "foo")
	}
}

func TestCrudUpdate(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	if !crud.Create(1, "foo") {
		t.Fatalf("got: false when calling Create, expected: true")
	}

	if !crud.Update(1, func(old string) string {
		if old != "foo" {
			t.Fatalf("got: %v as update func arg, expected: %v", old, "foo")
		}

		return "bar"
	}) {
		t.Fatalf("got: false when calling Update, expected: true")
	}

	value, ok := crud.Read(1)
	if !ok || value != "bar" {
		t.Fatalf("got: %v, %v when querying key, expected: %v, true", value, ok, "bar")
	}
}

func TestCrudUpdateAbsent(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	if crud.Update(1, func(old string) string { return "bar" }) {
		t.Fatalf("got: true when calling Update, expected: false")
	}

	value, ok := crud.Read(1)
	if ok || value != "" {
		t.Fatalf("got: %v, %v when querying key, expected: %v, false", value, ok, "")
	}
}

func TestCrudDelete(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	if !crud.Create(1, "foo") {
		t.Fatalf("got: false when calling Create, expected: true")
	}

	if !crud.Delete(1) {
		t.Fatalf("got: false when calling Delete, expected: true")
	}

	value, ok := crud.Read(1)
	if ok || value != "" {
		t.Fatalf("got: %v, %v when querying key, expected: %v, false", value, ok, "")
	}
}

func TestCrudDeleteAbsent(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	if crud.Delete(1) {
		t.Fatalf("got: true when calling Delete, expected: false")
	}
}
