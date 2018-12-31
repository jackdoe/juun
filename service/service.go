package main

import (
	"encoding/binary"
	"encoding/json"
	. "github.com/jackdoe/juun/common"
	. "github.com/jackdoe/juun/vw"
	"github.com/sevlyar/go-daemon"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
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
		log.Warnf("err: %s", err.Error())
		c.Close()
		return
	}
	ctrl := &Control{}
	err = json.Unmarshal(data, ctrl)
	if err != nil {
		log.Warnf("err: %s", err.Error())
		c.Close()
		return
	}

	out := ""
	log.Infof("datalen: %d %#v", dataLen, ctrl)
	switch ctrl.Command {
	case "add":
		if len(ctrl.Payload) > 0 {
			history.add(ctrl.Payload, ctrl.Pid, ctrl.Env)
		}

	case "end":
		history.gotoend(ctrl.Pid)

	case "delete":
		history.deletePID(ctrl.Pid)

	case "search":
		line := strings.Replace(ctrl.Payload, "\n", "", -1)
		if len(line) > 0 {
			out = history.search(line, ctrl.Pid, ctrl.Env)
		}
	case "up":
		out = history.up(ctrl.Pid, ctrl.Payload)
	case "down":
		out = history.down(ctrl.Pid, ctrl.Payload)
	}

	c.Write([]byte(out))

	c.Close()
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func listen(history *History, ln net.Listener) {
	for {
		fd, err := ln.Accept()
		if err != nil {
			log.Warnf("accept error: %s", err.Error())
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
func getHome() string {
	home := os.Getenv("HOME")
	if home == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		home = usr.HomeDir
	}
	return home
}

func main() {
	home := GetHome()
	histfile := path.Join(home, ".juun.json")
	socketPath := path.Join(home, ".juun.sock")
	pidFile := path.Join(home, ".juun.pid")
	configFile := path.Join(home, ".juun.config")
	modelFile := path.Join(home, ".juun.vw")
	if isRunning(pidFile) {
		os.Exit(0)
	}

	history := NewHistory()

	cntxt := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0600,
		LogFileName: path.Join(home, ".juun.log"),
		LogFilePerm: 0600,
		WorkDir:     home,
		Umask:       027,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		log.Fatal("Unable to run: ", err)
	}
	if d != nil {
		return
	}
	log.Infof("---------------------")
	log.Infof("loading %s, listening to: %s, model: %s", histfile, socketPath, modelFile)
	dat, err := ioutil.ReadFile(histfile)
	if err == nil {
		err = json.Unmarshal(dat, history)
		if err != nil {
			log.Warnf("err: %s", err.Error())
			history = NewHistory()
		}
	} else {
		log.Warnf("err: %s", err.Error())
	}

	history.selfReindex()

	config := NewConfig()
	dat, err = ioutil.ReadFile(configFile)
	if err == nil {
		err = json.Unmarshal(dat, config)
		if err != nil {
			config = NewConfig()
		}
		log.Infof("config[%s]: %s", configFile, prettyPrint(config))
	} else {
		log.Warnf("missing config file %s, using default: %s", configFile, prettyPrint(config))
	}

	if config.AutoSaveInteralSeconds < 30 {
		log.Warnf("autosave interval is too short, limiting it to 30 seconds")
		config.AutoSaveInteralSeconds = 30
	}
	level, err := log.ParseLevel(config.LogLevel)
	if err != nil {
		log.Warnf("failed to parse level %s: %s", config.LogLevel, err)
	} else {
		log.SetLevel(level)
	}
	log.SetReportCaller(true)

	var vw *Bandit
	if config.EnableVowpalWabbit {
		vw = NewBandit(modelFile) // XXX: can be nil if vw is not found
	}
	history.vw = vw
	syscall.Unlink(socketPath)
	sock, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatal("Listen error: ", err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	save := func() {
		history.lock.Lock()
		d1, err := json.Marshal(history)
		history.lock.Unlock()
		if err == nil {
			SafeSave(histfile, func(tmp string) error {
				return ioutil.WriteFile(tmp, d1, 0600)
			})
		} else {
			log.Warnf("error marshalling: %s", err.Error())
		}

		if vw != nil {
			vw.Save()
		}
	}

	cleanup := func() {
		log.Infof("closing")
		sock.Close()

		save()
		os.Chmod(modelFile, 0600)
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

	if config.AutoSaveInteralSeconds > 0 {
		go func() {
			for {
				save()
				time.Sleep(time.Duration(config.AutoSaveInteralSeconds) * time.Second)
			}
		}()
	}

	listen(history, sock)
	cleanup()
}
