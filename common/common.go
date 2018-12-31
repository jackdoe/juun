package common

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"os"
	"os/user"
	"strconv"
	"time"
)

func IntOrZero(s string) int {
	pid, _ := strconv.Atoi(s)
	return pid
}

type Control struct {
	Command string
	Payload string
	Pid     int
	Env     map[string]string
}

func GetOrDefault(env map[string]string, key string, def string) string {
	if env == nil {
		return def
	}
	v, ok := env[key]
	if !ok {
		return def
	}
	return v
}

func GetHome() string {
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

func GetCWD() string {
	dir, err := os.Getwd()
	if err != nil {
		dir = ""
	}
	return dir
}

func SafeSave(fn string, cb func(temp string) error) {
	tmp := fmt.Sprintf("%s.%s.tmp", fn, randSeq(10))
	defer os.Remove(tmp)
	log.Infof("saving %s", tmp)

	err := cb(tmp)

	if err != nil {
		log.Warnf("%s", err.Error())
	} else {
		log.Infof("renaming %s to %s", tmp, fn)
		err := os.Rename(tmp, fn)
		if err != nil {
			log.Warnf("%s", err.Error())
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
