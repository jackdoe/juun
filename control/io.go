package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/user"
	"path"
)

func query(cmd string, pid string, line string) string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	socketPath := path.Join(usr.HomeDir, ".juun.sock")
	c, err := net.Dial("unix", socketPath)
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
