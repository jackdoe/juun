package main

import (
	"log"
)

type Terminal struct {
	Commands                []int
	CommandsSet             map[int]bool
	Cursor                  int
	GlobalIdOnStart         int
	GlobalId                int
	CurrentBufferBeforeMove string
	globalMode              bool
	direction               int
}

const DIR_UP = 1
const DIR_DOWN = 2
const DIR_NONE = 3

func (t *Terminal) add(id int) {
	t.Commands = append(t.Commands, id)
	t.CommandsSet[id] = true
	t.end()
	t.globalMode = false
	t.direction = DIR_NONE
}

func (t *Terminal) up() (int, bool) {
	if t.isAtBeginning() {
		return 0, false
	}

	t.log("before up")
	defer t.log("  -> after up")
	wasDOWN := t.direction == DIR_DOWN
	t.direction = DIR_UP
	if wasDOWN {
		_, can := t.up() // bring it to the current level
		if !can {
			return 0, false
		}

		return t.up()
	}

	if !t.globalMode && t.Cursor <= len(t.Commands) {
		id := t.Commands[t.Cursor-1]
		if t.Cursor > 0 {
			t.Cursor--
		}

		return id, true
	} else {
		t.globalMode = true
		if t.GlobalId > 0 {
			id := t.GlobalId
			t.GlobalId--
			return id, true
		}

		return 0, false
	}
}

func (t *Terminal) down() (int, bool) {
	if t.isAtEnd() {
		return 0, false
	}
	t.log("before down")
	defer t.log("  -> after down")
	wasUP := t.direction == DIR_UP
	t.direction = DIR_DOWN
	if wasUP {
		_, can := t.down() // bring it to the current level
		if !can {
			return 0, false
		}
		return t.down()
	}
	if t.globalMode {
		if t.GlobalId <= t.GlobalIdOnStart {
			t.GlobalId++
			id := t.GlobalId
			return id, true
		} else {
			t.globalMode = false

			if len(t.Commands) > 0 {
				t.Cursor = 1
				return t.Commands[0], true
			}
			return 0, false
		}
	}

	if t.Cursor <= len(t.Commands)-1 {
		if t.Cursor < len(t.Commands) {
			t.Cursor++
		}

		id := t.Commands[t.Cursor-1]
		return id, true
	}

	return 0, false
}

func (t *Terminal) end() {
	t.GlobalId = t.GlobalIdOnStart
	t.Cursor = len(t.Commands)
}

func (t *Terminal) isAtEnd() bool {
	return t.GlobalId >= t.GlobalIdOnStart && t.Cursor == len(t.Commands)
}

func (t *Terminal) isAtBeginning() bool {
	return t.GlobalId == 0 && t.Cursor == 0
}

func (t *Terminal) log(p string) {
	log.Printf("%s: [global:%t] globalId: %d/%d, cursor:%d/%d, commands: %#v, buf: %s", p, t.globalMode, t.GlobalId, t.GlobalIdOnStart, t.Cursor, len(t.Commands)-1, t.Commands, t.CurrentBufferBeforeMove)
}
