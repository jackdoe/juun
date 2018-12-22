package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
)

const UNIX_SOCKET_PATH = "/tmp/juun.sock"

func query(cmd string, pid string, line string) string {
	c, err := net.Dial("unix", UNIX_SOCKET_PATH)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()
	_, err = c.Write([]byte(fmt.Sprintf("%s %s %s\n", cmd, pid, line)))
	if err != nil {
		log.Fatal("Write error:", err)
	}

	buf, _ := ioutil.ReadAll(c)
	return string(buf)
}

func main() {
	fmt.Printf("%s", query(os.Args[1], os.Args[2], os.Args[3]))
}
