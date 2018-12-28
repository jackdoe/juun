package main

import (
	"fmt"
	"github.com/chzyer/readline"
	. "github.com/jackdoe/juun/common"
	"os"
	"strings"
)

func main() {
	forceInterrupted := false
	cfg := &readline.Config{
		Prompt:            " \033[31m»\033[0m ",
		HistorySearchFold: false,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			switch r {
			case readline.CharCtrlZ:
				return r, false

			case readline.CharLineEnd:
				forceInterrupted = true
				return readline.CharInterrupt, true
			case readline.CharLineStart:
				forceInterrupted = true
				return readline.CharInterrupt, true

			case readline.CharBckSearch:
				return r, false
			case readline.CharFwdSearch:
				return r, false

			}

			return r, true
		},
		DisableAutoSaveHistory: true,
	}
	rl, err := readline.NewEx(cfg)
	if err != nil {
		return
	}
	defer rl.Close()

	result := ""

	cfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		if line != nil {
			result = QueryService("search", os.Args[1], string(line))
			rl.SetPrompt(fmt.Sprintf("%s \033[31m»\033[0m ", strings.Replace(result, "\n", "\\n", -1)))
		}
		rl.Refresh()
		return line, 0, false
	})

	cfg = rl.SetConfig(cfg)

	line, err := rl.Readline()
	rl.Clean()
	rl.SetPrompt("")
	exitCode := 0
	if forceInterrupted || err != nil {
		exitCode = 1
	}

	if !forceInterrupted && err != nil {
		fmt.Fprintf(rl.Stderr(), "")
	} else {
		if result == "" {
			fmt.Fprintf(rl.Stderr(), "%s", line)
		} else {
			fmt.Fprintf(rl.Stderr(), "%s", result)
		}
	}
	os.Exit(exitCode)
}
