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
	t.direction = DIR_NONE
}

// [ g 1 2 3 4 5 ]
//             +
// up
// [ g 1 2 3 4 5 ]
//           +
// down
// [ g 1 2 3 4 5 ]
//             +
func (t *Terminal) up() (int, bool) {
	wasDOWN := t.direction == DIR_DOWN
	t.direction = DIR_UP

	if len(t.Commands) == 0 {
		return 0, false
	}

	if t.Cursor > 0 {
		id := t.Commands[t.Cursor-1]
		t.Cursor--
		if wasDOWN && t.Cursor < len(t.Commands)-2 {
			id = t.Commands[t.Cursor+1]
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

// [ g 1 2 3 4 5 ]
//               +
// up
// [ g 1 2 3 4 5 ]
//             +
// down
// [ g 1 2 3 4 5 ]
//               +
func (t *Terminal) down() (int, bool) {
	if len(t.Commands) == 0 {
		return 0, false
	}
	wasUP := t.direction == DIR_UP
	t.direction = DIR_DOWN
	if t.globalMode {
		if t.GlobalId < t.GlobalIdOnStart {
			t.GlobalId++
			return t.GlobalId, true
		} else {
			t.globalMode = false
			return t.Commands[0], true
		}
	}

	if t.Cursor < len(t.Commands)-1 {
		t.Cursor++
		id := t.Commands[t.Cursor]

		if wasUP {
			id = t.Commands[t.Cursor]
		}
		return id, true
	}
	//	t.log("cant go down")
	return 0, false

}

func (t *Terminal) end() {
	t.Cursor = len(t.Commands)

}

func (t *Terminal) isAtEnd() bool {
	return t.Cursor >= len(t.Commands)
}

func (t *Terminal) log(p string) {
	log.Printf("%s: globalId: %d, cursor:%d, last index: %d, at end: %t, commands: %#v, buf: %s", p, t.GlobalId, t.Cursor, len(t.Commands)-1, t.isAtEnd(), t.Commands, t.CurrentBufferBeforeMove)
}
