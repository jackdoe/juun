package main

import "testing"
import "log"

func TestVW(t *testing.T) {
	v := NewVowpalInstance()

	v.SendReceive("1 |a b c")

	fs := NewFeatureSet(NewNamespace("a", NewFeature("abc", 0), NewFeature("abc", 1), NewFeature("|a^512653", 1)), NewNamespace("x", NewFeature("xyz", 0), NewFeature("xyz", 1)))
	log.Printf(fs.toVW)
	expected := "|a abc abc:1 _a_512653:1  |x xyz xyz:1  "
	if fs.toVW != expected {
		t.Fatalf("'%s' got '%s'", expected, fs.toVW)
	}

	log.Printf("%f", v.getVowpalScore("|a b 1"))
	v.Shutdown()

	bandit := NewBandit()

	pred := bandit.Predict(&item{id: 5, features: "|a b 1"}, &item{id: 6, features: "|a b 1"})
	if len(pred) != 2 {
		t.Fatalf("expected 2 items")
	}
	bandit.Click(5)
	bandit.Shutdown()

}
