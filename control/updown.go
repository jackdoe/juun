package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := ""
	if len(os.Args) > 3 {
		cmd = os.Args[3]
	}
	fmt.Printf("%s", query(os.Args[1], os.Args[2], cmd))
}
