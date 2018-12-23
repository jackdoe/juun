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
}

func (t *Terminal) currentCommandId() int {
	if len(t.Commands) == 0 {
		return 0
	}

	return t.Commands[t.Cursor]
}

func (t *Terminal) add(id int) {
	t.Commands = append(t.Commands, id)
	t.CommandsSet[id] = true
	t.end()
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

	if len(t.Commands) == 0 {
		return 0, false
	}

	if t.Cursor > 0 {
		id := t.Commands[t.Cursor-1]
		t.Cursor--
		return id, true
	} else {
		t.globalMode = true
		if t.GlobalId > 0 {
			id := t.GlobalId
			t.GlobalId--
			//			t.log("can go up")
			return id, true
		}

		return 0, false
	}
}

// [ g 1 2 3 4 5 ]
//             +
// up
// [ g 1 2 3 4 5 ]
//           +
// down
// [ g 1 2 3 4 5 ]
//             +
func (t *Terminal) down() (int, bool) {
	if len(t.Commands) == 0 {
		return 0, false
	}
	if t.globalMode {
		if t.GlobalId < t.GlobalIdOnStart {
			//			id := t.GlobalId
			t.GlobalId++
			return t.GlobalId, true
		} else {
			t.globalMode = false
			return t.Commands[0], true
		}
	}
	if t.Cursor < len(t.Commands)-1 {
		id := t.Commands[t.Cursor]
		t.Cursor++
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
