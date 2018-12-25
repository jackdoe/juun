package main

import "testing"

func TestVW(t *testing.T) {
	v := NewVowpalInstance()

	v.SendReceive("1 |a b c\n")

	v.Shutdown()
}
