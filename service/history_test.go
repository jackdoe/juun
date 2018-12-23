package main

import "testing"

func TestHistory(t *testing.T) {
	h := NewHistory()
	h.add("make", 1)
	h.add("make", 2)
	h.add("make", 2)

	if len(h.Lines) != 1 {
		t.Fatalf("expected only 1 line")
	}
	if h.search("m", 1) != "make" {
		t.Fatalf("make not found")
	}

	if h.search("make", 1) != "make" {
		t.Fatalf("make not found")
	}
}

func must(t *testing.T, a, b string) {
	if a != b {
		t.Logf("result: %s != expected: %s", a, b)
		panic("a")
	}
}

func TestHistoryChange(t *testing.T) {
	h := NewHistory()
	h.add("first-terminal-ps 1", 1) // global id 0
	h.add("ps 2", 2)                // global id 1
	h.add("ps 3", 2)                // global id 1

	must(t, h.up(2, "incomplete-before-up"), "ps 3")                // global id 1, cursor 2
	must(t, h.up(2, "incomplete-before-up"), "ps 2")                // global id 1, cursor 0
	must(t, h.up(2, "incomplete-before-up"), "first-terminal-ps 1") // global id 1, cursor 0
}

func TestHistoryUpDown(t *testing.T) {
	h := NewHistory()
	must(t, h.up(1, "incomplete"), "")
	must(t, h.down(1, ""), "incomplete")

	h.add("ps 1", 1)

	must(t, h.up(1, "incomplete"), "ps 1")
	must(t, h.down(1, ""), "incomplete")

	h.add("ps 2", 1)
	h.add("ps 3", 1)
	h.add("ps 4", 1)

	must(t, h.up(1, "incomplete-before-up"), "ps 4")
	for i := 0; i < 100; i++ {
		h.up(1, "")
	}
	must(t, h.down(1, ""), "ps 1")
	for i := 0; i < 100; i++ {
		h.down(1, "")
	}

	must(t, h.down(1, ""), "incomplete-before-up")
}

func TestHistoryUpDownMany(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", 1)
	h.add("ps 3", 1)
	h.add("ps 4", 1)

	must(t, h.up(1, "incomplete-before-up"), "ps 4")
	for i := 0; i < 100; i++ {
		h.up(1, "")
	}
	must(t, h.down(1, ""), "ps 2")
	for i := 0; i < 100; i++ {
		h.down(1, "")
	}

	must(t, h.down(1, ""), "incomplete-before-up")
}

func TestGlobalHistory(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", 1)
	h.add("ps 3", 1)
	h.add("ps 4", 1)

	h.add("zs 1", 2)
	h.add("zs 2", 2)

	h.add("zs 3", 2)
	h.add("zs x", 3)

	must(t, h.up(3, "incomplete-before-up"), "zs x")
	must(t, h.up(3, ""), "zs 3")
	for i := 0; i < 100; i++ {
		h.up(3, "")
	}
	must(t, h.up(3, ""), "ps 2")

	must(t, h.down(3, ""), "ps 3")
	must(t, h.down(3, ""), "ps 4")
	must(t, h.down(3, ""), "zs 1")
	must(t, h.down(3, ""), "zs 2")
	must(t, h.down(3, ""), "zs 3")
	must(t, h.down(3, ""), "zs x")
	must(t, h.down(3, ""), "incomplete-before-up")
}

func TestUpDownUp(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", 1)
	h.add("ps 2", 1)
	h.add("ps 3", 1)

	must(t, h.up(1, "incomplete-before-up"), "ps 3")
	must(t, h.up(1, ""), "ps 2")
	must(t, h.down(1, ""), "ps 3")
	must(t, h.down(1, ""), "incomplete-before-up")
	must(t, h.up(1, ""), "ps 3")
}
