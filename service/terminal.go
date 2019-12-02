package main

type Terminal struct {
	Commands    []int
	CommandsSet map[int]int

	Cursor int

	GlobalIdOnStart int

	CurrentBufferBeforeMove string
}

func NewTerminal(currentHistoryId int) *Terminal {
	return &Terminal{
		Commands:        []int{},
		Cursor:          0,
		GlobalIdOnStart: currentHistoryId,
		CommandsSet:     map[int]int{},
	}
}

func (t *Terminal) add(id int) {
	//	defer t.log("ADD")
	idx := len(t.Commands)
	t.Commands = append(t.Commands, id)
	t.CommandsSet[id] = idx
	t.end()
}

func (t *Terminal) up() (int, bool) {
	//	t.log("before up")
	//	defer t.log("  -> after up")
	if t.Cursor-1 <= -t.GlobalIdOnStart {
		if t.Cursor-1 == -t.GlobalIdOnStart {
			t.Cursor--
		}

		return 0, false
	}

	t.Cursor--
	if t.Cursor >= 0 {
		return t.Commands[t.Cursor], true
	}
	return t.GlobalIdOnStart + t.Cursor, true
}

func (t *Terminal) down() (int, bool) {
	//	t.log("before down")
	//	defer t.log("  -> after down")

	if t.Cursor+1 > len(t.Commands)-1 {
		if t.Cursor == len(t.Commands)-1 {
			t.Cursor++
		}

		return 0, false
	}

	t.Cursor++

	if t.Cursor >= 0 {
		return t.Commands[t.Cursor], true
	}
	return t.GlobalIdOnStart + t.Cursor, true
}

func (t *Terminal) end() {
	t.Cursor = len(t.Commands)
}

func (t *Terminal) isAtEnd() bool {
	return t.Cursor == len(t.Commands)
}
