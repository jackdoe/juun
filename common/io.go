package common

import (
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"path"
)

func QueryService(cmd string, spid string, line string) string {
	pid := IntOrZero(spid)
	ctrl := &Control{
		Command: cmd,
		Payload: line,
		Pid:     pid,
		Env: map[string]string{
			"cwd": GetCWD(),
		},
	}
	home := GetHome()
	data, err := json.Marshal(ctrl)
	if err != nil {
		log.Fatal("encoding error:", err)
	}

	socketPath := path.Join(home, ".juun.sock")
	c, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatal("Dial error", err)
	}
	defer c.Close()

	header := make([]byte, 4)
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
