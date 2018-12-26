package common

import (
	"log"
	"os"
	"os/user"
	"strconv"
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
