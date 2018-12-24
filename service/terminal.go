package main

import (
	"log"
)

type Terminal struct {
	Commands    []int
	CommandsSet map[int]bool

	Cursor     int
	CursorPrep int

	GlobalIdOnStart int
	GlobalId        int
	GlobalIdPrep    int

	CurrentBufferBeforeMove string

	globalMode bool
	direction  int
}

const DIR_UP = 1
const DIR_DOWN = 2
const DIR_NONE = 3

func (t *Terminal) prepare(p int) (int, bool) {
	if p > 0 && p <= len(t.Commands) {
		t.CursorPrep = p
		return t.Commands[p-1], true
	}

	return 0, false
}

func (t *Terminal) commit() {
	t.Cursor = t.CursorPrep
}

func (t *Terminal) prepareG(p int) (int, bool) {
	if p > 0 && p <= t.GlobalIdOnStart {
		t.GlobalIdPrep = p
		return p, true
	}
	return 0, false
}

func (t *Terminal) commitG() {
	t.GlobalId = t.GlobalIdPrep
}

func (t *Terminal) add(id int) {
	t.Commands = append(t.Commands, id)
	t.CommandsSet[id] = true
	t.end()
	t.globalMode = false
	t.direction = DIR_NONE
}

func (t *Terminal) up() (int, bool) {
	t.log("before up")
	defer t.log("  -> after up")

	wasUP := t.direction == DIR_UP
	if wasUP {
		t.commit()
		t.commitG()
	}
	t.direction = DIR_UP

	if !t.globalMode && t.Cursor <= len(t.Commands) {
		id := t.Commands[t.Cursor-1]
		t.prepare(t.Cursor - 1)
		return id, true
	} else {
		t.globalMode = true
		id := t.GlobalId
		t.prepareG(t.GlobalId - 1)
		return id, true
	}
}

func (t *Terminal) down() (int, bool) {
	t.log("before down")
	defer t.log("  -> after down")
	wasDOWN := t.direction == DIR_DOWN
	if wasDOWN {
		t.commit()
		t.commitG()
	}

	t.direction = DIR_DOWN
	if t.globalMode {
		id, can := t.prepare(t.GlobalId + 1)
		if !can {
			t.globalMode = false

			if len(t.Commands) > 0 {
				t.Cursor = 1
				return t.Commands[0], true
			}
			return 0, false
		} else {
			return id, can
		}
	}

	return t.prepare(t.Cursor + 1)
}

func (t *Terminal) end() {
	t.GlobalId = t.GlobalIdOnStart
	t.Cursor = len(t.Commands)
}

func (t *Terminal) isAtEnd() bool {
	return t.GlobalId == t.GlobalIdOnStart && t.Cursor == len(t.Commands)
}

func (t *Terminal) isAtBeginning() bool {
	return t.GlobalId == 0 && t.Cursor == 0
}

func (t *Terminal) log(p string) {
	log.Printf("%s: [global:%t] globalId: %d/%d, cursor:%d/%d, commands: %#v, buf: %s", p, t.globalMode, t.GlobalId, t.GlobalIdOnStart, t.Cursor, len(t.Commands)-1, t.Commands, t.CurrentBufferBeforeMove)
}
