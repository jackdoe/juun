package main

import (
	"fmt"
	. "github.com/jackdoe/juun/common"
	"os"
)

func main() {
	cmd := ""
	if len(os.Args) > 3 {
		cmd = os.Args[3]
	}
	fmt.Printf("%s", QueryService(os.Args[1], os.Args[2], cmd))
}
