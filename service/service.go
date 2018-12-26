package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/sevlyar/go-daemon"
	"io"
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
	"time"
)

func intOrZero(s string) int {
	pid, _ := strconv.Atoi(s)
	return pid
}

func oneLine(history *History, c net.Conn) {
	hdr := make([]byte, 4)
	_, err := io.ReadFull(c, hdr)
	if err != nil {
		c.Close()
		return
	}
	dataLen := binary.LittleEndian.Uint32(hdr)
	data := make([]byte, dataLen)
	_, err = io.ReadFull(c, data)
	if err != nil {
		c.Close()
		return
	}

	input := string(data)

	// cmd pid rest
	if err == nil {
		splitted := strings.SplitN(input, " ", 3)
		pid := intOrZero(splitted[1])
		line := splitted[2]
		out := ""
		log.Printf("datalen: %d pid: %d action %s line: %s", dataLen, pid, splitted[0], line)
		switch splitted[0] {
		case "add":
			if len(line) > 0 {
				history.add(line, pid)
			}

		case "end":
			history.gotoend(pid)

		case "delete":
			history.deletePID(pid)

		case "search":
			line = strings.Replace(line, "\n", "", -1)
			if len(line) > 0 {
				out = history.search(line, pid)
			}
		case "up":
			out = history.up(pid, line)
		case "down":
			out = history.down(pid, line)

		}
		c.Write([]byte(out))
	}
	c.Close()
}

func listen(history *History, ln net.Listener) {
	for {
		fd, err := ln.Accept()
		if err != nil {
			log.Print("accept error:", err)
			break
		}

		go oneLine(history, fd)
	}
}

func isRunning(pidFile string) bool {
	if piddata, err := ioutil.ReadFile(pidFile); err == nil {
		if pid, err := strconv.Atoi(string(piddata)); err == nil {
			if process, err := os.FindProcess(pid); err == nil {
				if err := process.Signal(syscall.Signal(0)); err == nil {
					return true
				}
			}
		}
	}
	return false
}

func main() {
	history := NewHistory()
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	histfile := path.Join(usr.HomeDir, ".juun.json")
	socketPath := path.Join(usr.HomeDir, ".juun.sock")
	pidFile := path.Join(usr.HomeDir, ".juun.pid")
	modelFile := path.Join(usr.HomeDir, ".juun.vw")
	if isRunning(pidFile) {
		os.Exit(0)
	}

	cntxt := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0600,
		LogFileName: path.Join(usr.HomeDir, ".juun.log"),
		LogFilePerm: 0600,
		WorkDir:     usr.HomeDir,
		Umask:       027,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	log.Printf("loading %s, listening to: %s, model: %s", histfile, socketPath, modelFile)
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

	history.selfReindex()

	vw := NewBandit(modelFile) // XXX: can be nil if vw is not found
	history.vw = vw
	syscall.Unlink(socketPath)
	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	save := func() {

		d1, err := json.Marshal(history)
		tmp := fmt.Sprintf("%s.tmp", histfile)
		log.Printf("saving %s", tmp)
		if err == nil {
			err := ioutil.WriteFile(tmp, d1, 0600)
			if err != nil {
				log.Printf("%s", err.Error())
			} else {
				log.Printf("renaming %s to %s", tmp, histfile)
				err := os.Rename(tmp, histfile)
				if err != nil {
					log.Printf("%s", err.Error())
				}
			}
		} else {
			log.Printf("%s", err.Error())
		}
		if vw != nil {
			vw.Save()
		}
	}

	cleanup := func() {
		log.Printf("closing")
		save()
		os.Chmod(modelFile, 0600)
		sock.Close()
		if vw != nil {
			vw.Shutdown()
		}
		cntxt.Release()
		os.Exit(0)
	}

	go func() {
		<-sigs
		cleanup()
	}()

	go func() {
		for {
			save()
			time.Sleep(300 * time.Second)
		}
	}()

	listen(history, sock)
	cleanup()
}
