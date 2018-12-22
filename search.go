package main

import (
	"fmt"
	"github.com/chzyer/readline"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
)

const UNIX_SOCKET_PATH = "/tmp/juun.sock"

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
func query(pid string, line string) string {
	c, err := net.Dial("unix", UNIX_SOCKET_PATH)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	_, err = c.Write([]byte(fmt.Sprintf("search %s %s\n", pid, line)))
	if err != nil {
		log.Fatal("Write error:", err)
	}

	buf, _ := ioutil.ReadAll(c)
	return string(buf)
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
			result = query(os.Args[1], string(line))
			rl.SetPrompt(fmt.Sprintf("%s \033[31m»\033[0m ", strings.Replace(result, "\n", "", -1)))
		}
		rl.Refresh()
		return line, 0, false
	})

	cfg = rl.SetConfig(cfg)

	line, err := rl.Readline()

	if result == "" {
		fmt.Fprintf(rl.Stderr(), "\n%s\n", line)
	} else {
		fmt.Fprintf(rl.Stderr(), "\n%s\n", result)
	}
}
