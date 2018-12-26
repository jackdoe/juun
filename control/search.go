package main

import (
	"fmt"
	"github.com/chzyer/readline"
	. "github.com/jackdoe/juun/common"
	"os"
	"strings"
)

func filterInput(r rune) (rune, bool) {
	switch r {
	case readline.CharCtrlZ:
		return r, false
	case readline.CharBckSearch:
		return r, false
	}
	return r, true
}

func main() {
	cfg := &readline.Config{
		Prompt:                 " \033[31m»\033[0m ",
		HistorySearchFold:      false,
		FuncFilterInputRune:    filterInput,
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

	if err != nil || result == "" {
		if result == "" {
			fmt.Fprintf(rl.Stderr(), "%s", line)
		} else {
			fmt.Fprintf(rl.Stderr(), "%s", result)
		}
		os.Exit(1)
	} else {
		fmt.Fprintf(rl.Stderr(), "%s", result)
		os.Exit(0)
	}
}
