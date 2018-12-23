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

func (t *Terminal) canUP() bool {
	return !t.isAtBeginning()
}

func (t *Terminal) canDOWN() bool {
	return !t.isAtEnd()
}
func (t *Terminal) inc() {
	if t.Cursor < len(t.Commands) {
		t.Cursor++
	}
}
func (t *Terminal) dec() {
	if t.Cursor > 0 {
		t.Cursor--
	}
}

func (t *Terminal) up() (int, bool) {
	if t.isAtBeginning() {
		return 0, false
	}

	t.log("before up")
	defer t.log("  -> after up")

	if t.Cursor < len(t.Commands) {
		id := t.Commands[t.Cursor-1]
		t.dec()
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
	delta := 1
	if wasUP && !t.isAtBeginning() {
		delta = 2
	}
	t.direction = DIR_DOWN
	if t.globalMode {
		if t.GlobalId+delta <= t.GlobalIdOnStart {
			t.GlobalId += delta
			id := t.GlobalId
			return id, true
		} else {
			t.globalMode = false

			if len(t.Commands) > 0 {
				t.Cursor = delta
				return t.Commands[0], true
			}
			return 0, false
		}
	}

	if t.Cursor <= len(t.Commands)-delta {
		t.Cursor += delta
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
	log.Printf("%s: [global:%t] globalId: %d/%d, cursor:%d/%d, at end/up/down: %t/%t/%t, commands: %#v, buf: %s", p, t.globalMode, t.GlobalId, t.GlobalIdOnStart, t.Cursor, len(t.Commands)-1, t.isAtEnd(), t.canUP(), t.canDOWN(), t.Commands, t.CurrentBufferBeforeMove)
}
