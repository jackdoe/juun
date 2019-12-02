package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/chzyer/readline"
	. "github.com/jackdoe/juun/common"
)

func main() {
	result := []*HistoryLine{}
	lastQuery := ""
	currentIndex := 0

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

	cfg.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		if line != nil {
			currentQuery := string(line)
			if strings.Trim(lastQuery, " ") == strings.Trim(currentQuery, " ") && len(result) > 0 {
				currentIndex++
				r := result[currentIndex%len(result)]
				rl.SetPrompt(fmt.Sprintf("%s \033[95mϵ\033[0m %d/%d \033[31m»\033[0m ", strings.Replace(r.Line, "\n", "\\n", -1), currentIndex%len(result), len(result)))
			} else {
				encoded := QueryService("search", os.Args[1], currentQuery)

				err := json.Unmarshal([]byte(encoded), &result)
				if err == nil {
					if len(result) > 0 {
						lastQuery = currentQuery
						currentIndex = 0
						r := result[0]
						rl.SetPrompt(fmt.Sprintf("%s \033[95mϵ\033[0m %d/%d \033[31m»\033[0m ", strings.Replace(r.Line, "\n", "\\n", -1), currentIndex%len(result), len(result)))
						lastQuery = currentQuery
					} else {
						rl.SetPrompt(fmt.Sprintf("%s \033[31m»\033[0m ", currentQuery))
					}
				}
			}
		}
		rl.Refresh()
		return line, 0, false
	})

	cfg = rl.SetConfig(cfg)

	_, err = rl.Readline()
	rl.Clean()
	rl.SetPrompt("")
	exitCode := 0
	if forceInterrupted || err != nil {
		exitCode = 1
	}

	if !forceInterrupted && err != nil {
		fmt.Fprintf(rl.Stderr(), "")
	} else {
		if len(result) > 0 {
			fmt.Fprintf(rl.Stderr(), "%s", result[currentIndex%len(result)].Line)
		} else {
			fmt.Fprintf(rl.Stderr(), "%s", result)
		}
	}
	os.Exit(exitCode)
}
