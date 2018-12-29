package main

import (
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	got := timeToNamespace("i_abc", time.Unix(0, 0)).ToVW()
	expected := "|i_abc year_1970 day_1 month_1 hour_1 "
	if got != expected {
		t.Fatalf("wrong time features, expecte: '%s', got '%s'", expected, got)
	}
}

func TestLimit(t *testing.T) {
	h := NewHistory()
	h.limit(2)
	h = NewHistory()

	h.add("ps 1", 1, nil)
	h.limit(1)
	if len(h.Lines) != 1 {
		t.Fatalf("%d != 1", len(h.Lines))
	}

	h.add("ps 1", 1, nil)
	h.add("ps 2", 1, nil)
	h.add("ps 3", 1, nil)
	if len(h.Lines) != 3 {
		t.Fatalf("%d != 2", len(h.Lines))
	}

	h.limit(2)
	if len(h.Lines) != 2 {
		t.Fatalf("%d != 2", len(h.Lines))
	}

	h.limit(2)
	if len(h.Lines) != 2 {
		t.Fatalf("%d != 2", len(h.Lines))
	}

}

func TestRemove(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", 1, nil)
	h.add("ps 2", 1, nil)
	h.add("ps 3", 1, nil)

	must(t, h.up(1, "incomplete-before-up"), "ps 3") // "" -> ps 3
	must(t, h.up(1, ""), "ps 2")                     // ps3 -> ps 2
	must(t, h.down(1, ""), "ps 3")                   // ps 2 -> ps 3
	must(t, h.down(1, ""), "incomplete-before-up")
	must(t, h.up(1, ""), "ps 3")

	h.removeLine("ps 2")
	if h.search("ps 1", 1, nil) != "ps 1" {
		t.Fatalf("ps 1 not found")
	}

	if h.search("ps 2", 1, nil) != "" {
		t.Fatalf("ps 2 found")
	}

	must(t, h.up(1, ""), "ps 3")
	must(t, h.up(1, ""), "ps 1")
	h.removeLine("ps 3")
	must(t, h.up(1, ""), "ps 1")
	h.removeLine("ps 1")
	must(t, h.up(1, ""), "")

	h.add("ps 3", 1, nil)
	if h.search("ps 3", 1, nil) != "ps 3" {
		t.Fatalf("ps 3 not found")
	}

	must(t, h.up(1, ""), "ps 3")
	h.add("ps 4", 1, nil)
	must(t, h.up(1, ""), "ps 4")

}

func TestUpDownUp(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", 1, nil)
	h.add("ps 2", 1, nil)
	h.add("ps 3", 1, nil)

	must(t, h.up(1, "incomplete-before-up"), "ps 3") // "" -> ps 3
	must(t, h.up(1, ""), "ps 2")                     // ps3 -> ps 2
	must(t, h.down(1, ""), "ps 3")                   // ps 2 -> ps 3
	must(t, h.down(1, ""), "incomplete-before-up")
	must(t, h.up(1, ""), "ps 3")
}

func TestHistory(t *testing.T) {
	h := NewHistory()
	h.add("make", 1, nil)
	h.add("make", 2, nil)
	h.add("make", 2, nil)

	if len(h.Lines) != 1 {
		t.Fatalf("expected only 1 line")
	}
	if h.search("m", 1, nil) != "make" {
		t.Fatalf("make not found")
	}

	if h.search("make", 1, nil) != "make" {
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
	h.add("first-terminal-ps 1", 1, nil)                            // global id 0
	h.add("ps 2", 2, nil)                                           // global id 1
	h.add("ps 3", 2, nil)                                           // global id 1
	must(t, h.up(2, "incomplete-before-up"), "ps 3")                // global id 1, cursor 2
	must(t, h.up(2, "incomplete-before-up"), "ps 2")                // global id 1, cursor 0
	must(t, h.up(2, "incomplete-before-up"), "first-terminal-ps 1") // global id 1, cursor 0
}

func TestGlobalHistory2(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", 1, nil) // -> 0
	h.add("ps 3", 1, nil) // -> 1
	h.add("ps 4", 1, nil) // -> 2

	h.add("zs 1", 2, nil)
	h.add("zs 2", 2, nil)
	h.add("zs 3", 2, nil)

	must(t, h.up(3, "incomplete-before-up"), "zs 3")
	must(t, h.up(3, ""), "zs 2")
	for i := 0; i < 100; i++ {
		h.up(3, "")
	}
	must(t, h.up(3, ""), "ps 2")
	must(t, h.down(3, ""), "ps 3")
	must(t, h.down(3, ""), "ps 4")
	must(t, h.down(3, ""), "zs 1")
	must(t, h.down(3, ""), "zs 2")
	must(t, h.down(3, ""), "zs 3")
	must(t, h.down(3, ""), "incomplete-before-up")
}

func TestGlobalHistory(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", 1, nil)
	h.add("ps 3", 1, nil)
	h.add("ps 4", 1, nil)

	h.add("zs 1", 2, nil)
	h.add("zs 2", 2, nil)

	h.add("zs 3", 2, nil)
	h.add("zs x", 3, nil)

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

func TestUpDownUpGlobal(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", 2, nil) //0
	h.add("ps 2", 2, nil) //1
	h.add("ps 3", 2, nil) //2

	must(t, h.up(1, "incomplete-before-up"), "ps 3") // 2 -> 1
	must(t, h.up(1, ""), "ps 2")                     // 1 -> 0
	must(t, h.down(1, ""), "ps 3")                   // 0 -> 1
	must(t, h.down(1, ""), "incomplete-before-up")   // 1 -> 2

	must(t, h.down(1, ""), "incomplete-before-up")
	must(t, h.up(1, ""), "ps 3")
}

func TestHistoryUpDown(t *testing.T) {
	h := NewHistory()

	must(t, h.up(1, "incomplete"), "")
	must(t, h.down(1, ""), "incomplete")

	h.add("ps 1", 1, nil)

	must(t, h.up(1, "incomplete"), "ps 1")
	must(t, h.down(1, ""), "incomplete")

	h.add("ps 2", 1, nil)
	h.add("ps 3", 1, nil)
	h.add("ps 4", 1, nil)

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

func TestUpDownUpUpGlobal(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", 2, nil) //0
	must(t, h.up(1, "a"), "ps 1")
	must(t, h.down(1, ""), "a")
	must(t, h.up(1, "a"), "ps 1")
	must(t, h.down(1, ""), "a")
	must(t, h.up(1, "a"), "ps 1")
	must(t, h.down(1, ""), "a")
	h.add("ps 2", 2, nil) //0
	must(t, h.up(1, "a"), "ps 1")
	must(t, h.down(1, ""), "a")
}

func TestUpDownUpLocal(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", 1, nil) //0
	h.add("ps 2", 1, nil) //1
	h.add("ps 3", 1, nil) //2

	must(t, h.up(1, "incomplete-before-up"), "ps 3") // 2 -> 1
	must(t, h.up(1, ""), "ps 2")                     // 1 -> 0
	must(t, h.down(1, ""), "ps 3")                   // 0 -> 1

	must(t, h.down(1, ""), "incomplete-before-up")
	must(t, h.up(1, ""), "ps 3")
}

func TestHistoryUpDownMany(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", 1, nil)
	h.add("ps 3", 1, nil)
	h.add("ps 4", 1, nil)

	must(t, h.up(1, "incomplete-before-up"), "ps 4")
	for i := 0; i < 10; i++ {
		h.up(1, "")
	}

	must(t, h.down(1, ""), "ps 3")

	for i := 0; i < 10; i++ {
		h.down(1, "")
	}

	must(t, h.down(1, ""), "incomplete-before-up")
}
