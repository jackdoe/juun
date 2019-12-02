package common

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net"
	"path"
	"strings"
	"time"

	. "github.com/jackdoe/juun/vw"
)

type HistoryLine struct {
	Line      string
	TimeStamp int64
	Count     uint32
	Id        int
}

func (l *HistoryLine) IndexableFields() map[string]string {
	return map[string]string{"line": l.Line}
}

func (l *HistoryLine) Featurize() *FeatureSet {
	features := []*Feature{}
	for _, s := range strings.Split(l.Line, " ") {
		features = append(features, NewFeature(s, 0))
	}
	text := NewNamespace("i_text", features...)

	count := NewNamespace("i_count", NewFeature("count", float32(math.Log(float64(1)+float64(l.Count)))))
	t := TimeToNamespace("i_time", time.Unix(l.TimeStamp/1000000000, 0))
	id := NewNamespace("i_id", NewFeature(fmt.Sprintf("id=%d", l.Id), float32(0)))

	return NewFeatureSet(id, text, count, t)
}

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

func TimeToNamespace(ns string, now time.Time) *Namespace {
	features := []*Feature{}
	hr, _, _ := now.Clock()

	features = append(features, NewFeature(fmt.Sprintf("year=%d", now.Year()), 0))
	features = append(features, NewFeature(fmt.Sprintf("day=%d", now.Day()), 0))
	features = append(features, NewFeature(fmt.Sprintf("month=%d", now.Month()), 0))
	features = append(features, NewFeature(fmt.Sprintf("hour=%d", hr), 0))

	return NewNamespace(ns, features...)
}

func UserContext(query string, cwd string) *FeatureSet {
	features := []*Feature{}
	for _, s := range strings.Split(query, " ") {
		if len(s) > 0 {
			features = append(features, NewFeature(s, 0))
		}
	}
	qns := NewNamespace("c_query", features...)

	fs := NewFeatureSet(TimeToNamespace("c_user_time", time.Now()), qns)
	if cwd != "" {
		splitted := strings.Split(cwd, "/")
		features := []*Feature{}
		if len(splitted) > 0 {
			features = append(features, NewFeature(splitted[len(splitted)-1], 0))
		}
		features = append(features, NewFeature(cwd, 0))
		fs.AddNamespaces(NewNamespace("c_cwd", features...))
	}
	return fs
}
