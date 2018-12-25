package main

import "testing"
import "log"

func TestVW(t *testing.T) {
	conn, cmd := startVW()

	sendReceive(conn, "1 |a b c\n")

	if err := cmd.Process.Kill(); err != nil {
		log.Fatal("failed to kill process: ", err)
	}
}
