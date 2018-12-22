package main

type Terminal struct {
	Commands        []int
	CommandsSet     map[int]bool
	Cursor          int
	GlobalIdOnStart int
	GlobalId        int
	globalMode      bool
}

func (t *Terminal) currentCommandId() int {
	if !t.globalMode {
		return t.Commands[t.Cursor]
	}
	return t.GlobalId
}

func (t *Terminal) up() {
	//	old := t.Cursor
	if t.Cursor > 0 {
		t.Cursor--
	} else {
		if !t.globalMode {
			//			log.Printf("enabling global mode")
			t.globalMode = true
		}
		if t.GlobalId > 0 {
			t.GlobalId--
		}
	}
	//	log.Printf("DOWN from %d to %d", old, t.Cursor)
}

func (t *Terminal) down() {
	//	old := t.Cursor

	if t.globalMode {
		if t.GlobalId >= t.GlobalIdOnStart {
			t.globalMode = false
			//			log.Printf("disabling global mode")
		} else {
			t.GlobalId++
		}
	} else {
		if t.Cursor < len(t.Commands)-1 {
			t.Cursor++
		}
	}
	//	log.Printf("UP from %d to %d", old, t.Cursor)
}
