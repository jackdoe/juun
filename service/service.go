package main

import (
	"bufio"
	"encoding/json"
	//	"github.com/takama/daemon"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"strings"
	"syscall"
)

func intOrZero(s string) int {
	pid, _ := strconv.Atoi(s)
	return pid
}

func oneLine(history *History, c net.Conn) {
	input, err := bufio.NewReader(c).ReadString('\n')
	// cmd pid rest
	log.Printf("input: %s", strings.Replace(input, "\n", " ", -1))
	if err == nil {
		splitted := strings.SplitN(input, " ", 3)
		pid := intOrZero(splitted[1])
		line := splitted[2]
		out := ""
		switch splitted[0] {
		case "add":
			if len(line) > 0 {
				history.add(line, pid)
			}

		case "search":
			if len(line) > 0 {
				out = history.search(strings.Trim(line, "\n"), pid)
			}
		case "up":
			out = history.move(true, pid)
		case "down":
			out = history.move(false, pid)

		}
		c.Write([]byte(out))
	}
	c.Close()
}

func listen(history *History, ln net.Listener) {
	for {
		fd, err := ln.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}

		go oneLine(history, fd)
	}
}

func main() {
	history := NewHistory()
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	histfile := path.Join(usr.HomeDir, ".juun.json")
	socketPath := path.Join(usr.HomeDir, ".juun.sock")

	dat, err := ioutil.ReadFile(histfile)
	if err == nil {
		err = json.Unmarshal(dat, history)
		if err != nil {
			log.Printf("err: %s", err.Error())
			history = NewHistory()
		}
	} else {
		log.Printf("err: %s", err.Error())
	}
	log.Printf("loading %s, listening to: %s", histfile, socketPath)

	syscall.Unlink(socketPath)
	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		log.Printf("closing")
		d1, err := json.Marshal(history)
		if err == nil {
			err := ioutil.WriteFile(histfile, d1, 0600)
			if err != nil {
				log.Printf("%s", err.Error())
			}
		} else {
			log.Printf("%s", err.Error())
		}
		sock.Close()
		os.Exit(0)
	}()

	listen(history, sock)
}
