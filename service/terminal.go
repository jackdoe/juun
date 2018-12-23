package main

import (
	"log"
)

type Terminal struct {
	Commands        []int
	CommandsSet     map[int]bool
	Cursor          int
	GlobalIdOnStart int
	GlobalId        int
}

func (t *Terminal) currentCommandId() int {
	// no commands yet on this terminal, just return the global id
	if len(t.Commands) == 0 {
		return t.GlobalId
	}

	if t.Cursor >= 0 {
		return t.Commands[t.Cursor]
	}
	return t.GlobalId
}

func (t *Terminal) up() bool {
	old := t.Cursor
	success := false
	if t.Cursor >= 0 {
		t.Cursor--
		success = true
	} else {
		if t.GlobalId > 0 {
			t.GlobalId--
			success = true
		}

	}

	log.Printf("UP from %d to %d global id %d current id: %d", old, t.Cursor, t.GlobalId, t.currentCommandId())
	return success
}

func (t *Terminal) down() bool {
	old := t.Cursor
	success := false
	if t.Cursor < 0 || len(t.Commands) == 0 {
		if t.GlobalId >= t.GlobalIdOnStart {
			t.Cursor = 0
			success = true
		} else {
			t.GlobalId++
			success = true
		}

	} else {
		if t.Cursor < len(t.Commands)-1 {
			t.Cursor++
			success = true
		}
	}
	log.Printf("DOWN from %d to %d global id %d current id: %d", old, t.Cursor, t.GlobalId, t.currentCommandId())
	return success
}

func (t *Terminal) end() {
	t.Cursor = len(t.Commands)
}
