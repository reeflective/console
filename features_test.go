package console

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

func TestHandleInterruptMatching(t *testing.T) {
	c := New("test")
	m := c.ActiveMenu()

	var fired []string
	sentinel := errors.New("boom")

	m.AddInterrupt(sentinel, func(*Console) { fired = append(fired, "sentinel") })
	m.AddInterrupt(io.EOF, func(*Console) { fired = append(fired, "eof") })

	// errors.Is match: a wrapped io.EOF should reach the io.EOF handler.
	fired = nil
	m.handleInterrupt(fmt.Errorf("read failed: %w", io.EOF))
	if !reflect_equal(fired, []string{"eof"}) {
		t.Fatalf("wrapped io.EOF fired %v, want [eof]", fired)
	}

	// String fallback: a distinct error value with the same message as the
	// registered sentinel should still match (the historical pattern).
	fired = nil
	m.handleInterrupt(errors.New("boom"))
	if !reflect_equal(fired, []string{"sentinel"}) {
		t.Fatalf("same-message error fired %v, want [sentinel]", fired)
	}

	// No match: nothing fires.
	fired = nil
	m.handleInterrupt(errors.New("unrelated"))
	if len(fired) != 0 {
		t.Fatalf("unrelated error fired %v, want none", fired)
	}
}

func TestMenuNewlineOverrides(t *testing.T) {
	c := New("test")
	c.NewlineAfter = true
	c.NewlineBefore = false
	c.NewlineWhenEmpty = false
	m := c.ActiveMenu()

	// With no override, the menu inherits the console defaults.
	if !m.newlineAfter() {
		t.Fatal("newlineAfter: expected inherited true")
	}
	if m.newlineBefore() {
		t.Fatal("newlineBefore: expected inherited false")
	}
	if m.newlineWhenEmpty() {
		t.Fatal("newlineWhenEmpty: expected inherited false")
	}

	// Overrides take precedence over the console default.
	m.SetNewlineAfter(false)
	m.SetNewlineBefore(true)
	m.SetNewlineWhenEmpty(true)

	if m.newlineAfter() {
		t.Fatal("newlineAfter: expected override false")
	}
	if !m.newlineBefore() {
		t.Fatal("newlineBefore: expected override true")
	}
	if !m.newlineWhenEmpty() {
		t.Fatal("newlineWhenEmpty: expected override true")
	}

	// Changing the console default no longer affects an overridden menu.
	c.NewlineAfter = true
	if m.newlineAfter() {
		t.Fatal("newlineAfter: override should shadow console default")
	}
}

func TestMenuEmptyCharsOverride(t *testing.T) {
	c := New("test")
	m := c.ActiveMenu()

	// Inherits the console default set.
	if string(m.emptyCharSet()) != string(c.EmptyChars) {
		t.Fatalf("emptyCharSet inherited = %q, want %q", string(m.emptyCharSet()), string(c.EmptyChars))
	}

	// Override.
	m.SetEmptyChars('x', 'y')
	if string(m.emptyCharSet()) != "xy" {
		t.Fatalf("emptyCharSet override = %q, want %q", string(m.emptyCharSet()), "xy")
	}

	// No arguments clears the override, restoring inheritance.
	m.SetEmptyChars()
	if string(m.emptyCharSet()) != string(c.EmptyChars) {
		t.Fatalf("emptyCharSet after clear = %q, want %q", string(m.emptyCharSet()), string(c.EmptyChars))
	}
}

func TestConsoleDefaultSignals(t *testing.T) {
	c := New("test")
	if len(c.Signals) != len(defaultTrapSignals) {
		t.Fatalf("default Signals = %v, want %v", c.Signals, defaultTrapSignals)
	}
}

// reflect_equal is a tiny string-slice comparison helper to avoid importing
// reflect for a single use.
func reflect_equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
