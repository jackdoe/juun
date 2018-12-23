package main

import (
	"encoding/binary"
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
	header := make([]byte, 4)
	data := []byte(fmt.Sprintf("%s %s %s", cmd, pid, line))
	binary.LittleEndian.PutUint32(header, uint32(len(data)))

	_, err = c.Write(header)
	if err != nil {
		log.Fatal("Write error:", err)
	}

	_, err = c.Write(data)
	if err != nil {
		log.Fatal("Write error:", err)
	}
	buf, _ := ioutil.ReadAll(c)
	return string(buf)
}
