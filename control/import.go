package main

import (
	"bufio"
	. "github.com/jackdoe/juun/common"
	"log"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	pid := "0"

	for scanner.Scan() {
		splitted := strings.SplitN(strings.TrimLeft(scanner.Text(), " "), "  ", 2)
		if len(splitted) == 2 {
			s := strings.Replace(splitted[1], "\\n", "\n", -1)
			log.Printf("adding %s", s)
			QueryService("add", pid, s)
		}
	}

	if scanner.Err() != nil {
		log.Printf("err: %s", scanner.Err())
	}
	QueryService("delete", pid, "delete")
}
