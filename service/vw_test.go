package main

import "testing"
import "log"
import "os"
import "time"

func TestVW(t *testing.T) {
	fn := "/tmp/juun-testing.vw"
	os.Remove(fn)
	v := NewVowpalInstance(fn)
	v.SendReceive("1 |a b c")

	log.Printf("%#v", v.SendReceive("| 0:"))
	v.Save()
	if !exists(fn) {
		t.Fatalf("missing %s", fn)
	}
	fs := NewFeatureSet(NewNamespace("a", NewFeature("abc", 0), NewFeature("abc", 1), NewFeature("|a^512653", 1)), NewNamespace("x", NewFeature("xyz", 0), NewFeature("xyz", 1)))
	log.Printf(fs.toVW())
	expected := "|a abc abc:1.000000 _a_512653:1.000000  |x xyz xyz:1.000000  "
	if fs.toVW() != expected {
		t.Fatalf("'%s' got '%s'", expected, fs.toVW())
	}

	log.Printf("%f", v.getVowpalScore("|a b 1"))

	v.Shutdown()
	os.Remove("/tmp/juun-testing.bandit.vw")
	bandit := NewBandit("/tmp/juun-testing.bandit.vw")

	pred := bandit.Predict(2, &item{id: 5, features: "|a b 1.000000"}, &item{id: 6, features: "|a b 1.000000"})
	if len(pred) != 2 {
		t.Fatalf("expected 2 items")
	}

	pred = bandit.Predict(1, &item{id: 5, features: "|a b 1.000000"}, &item{id: 6, features: "|a b 1.000000"})
	if len(pred) != 2 {
		t.Fatalf("expected 2 items")
	}

	bandit.Click(5)
	bandit.Shutdown()

	got := timeToNamespace("i_abc", time.Unix(0, 0)).toVW()
	expected = "|i_abc year_1970 day_1 month_1 hour_1 "
	if got != expected {
		t.Fatalf("wrong time features, expecte: '%s', got '%s'", expected, got)
	}
}
