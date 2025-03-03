package crud_test

import (
	"testing"

	"github.com/fourls/soko/internal/crud"
)

func TestCrudReadAbsent(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	value := crud.Read(1)
	if value != "" {
		t.Fatalf("got: %v when querying key, expected: %v", value, "")
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
		actual := crud.Read(c.key)
		if actual != c.expected {
			t.Fatalf("got: %v for key %d, expected: %v", actual, c.key, c.expected)
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

	value := crud.Read(1)
	if value != "foo" {
		t.Fatalf("got: %v when querying key, expected: %v", value, "foo")
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

	value := crud.Read(1)
	if value != "bar" {
		t.Fatalf("got: %v when querying key, expected: %v", value, "bar")
	}
}

func TestCrudUpdateAbsent(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	if crud.Update(1, func(old string) string { return "bar" }) {
		t.Fatalf("got: true when calling Update, expected: false")
	}

	value := crud.Read(1)
	if value != "" {
		t.Fatalf("got: %v when querying key, expected: %v", value, "")
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

	value := crud.Read(1)
	if value != "" {
		t.Fatalf("got: %v when querying key, expected: %v", value, "")
	}
}

func TestCrudDeleteAbsent(t *testing.T) {
	crud := crud.New[int, string]()
	defer crud.Close()

	if crud.Delete(1) {
		t.Fatalf("got: true when calling Delete, expected: false")
	}
}
