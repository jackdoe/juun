package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Printf("%s", query(os.Args[1], os.Args[2], os.Args[3]))
}
