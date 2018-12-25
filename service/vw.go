package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type vowpal struct {
	conn      net.Conn
	cmd       *exec.Cmd
	rw        *bufio.ReadWriter
	fn        string
	modelPath string
}

func (v *vowpal) Shutdown() {
	v.conn.Close()
	log.Printf("removing %s", v.fn)

	cmd := v.cmd
	err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	if err != nil {
		log.Printf("kill failed: %v\n", err)
	}
	os.Remove(v.fn)
}

var CLEANUP_REG = regexp.MustCompile("[^a-zA-Z0-9]+")

type feature struct {
	feature string
	value   float32
}

func NewFeature(f string, value float32) *feature {
	clean := CLEANUP_REG.ReplaceAllString(f, "_")
	if len(clean) == 0 {
		clean = "_"
	}
	return &feature{
		feature: clean,
		value:   value,
	}
}

func (f *feature) toVW() string {
	if f.value != 0 {
		return fmt.Sprintf("%s:%.0f", f.feature, f.value)
	}
	return fmt.Sprintf("%s", f.feature)
}

type namespace struct {
	ns       string
	features []*feature
}

func NewNamespace(ns string, fs ...*feature) *namespace {
	return &namespace{
		ns:       ns,
		features: fs,
	}
}

func (n *namespace) toVW() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("|%s ", n.ns))
	for _, f := range n.features {
		sb.WriteString(f.toVW())
		sb.WriteString(" ")
	}
	return sb.String()
}

type featureSet struct {
	namespaces []*namespace
	toVW       string
}

func NewFeatureSet(nss ...*namespace) *featureSet {
	var sb strings.Builder
	for _, f := range nss {
		sb.WriteString(f.toVW())
		sb.WriteString(" ")
	}

	return &featureSet{
		namespaces: nss,
		toVW:       sb.String(),
	}
}

func (v *vowpal) Save() {
	v.rw.Write([]byte(fmt.Sprintf("save_%s", v.modelPath)))
	v.rw.Write([]byte("\n"))
	v.rw.Flush()
	waitForFile(v.modelPath)
}

func (v *vowpal) SendReceive(line string) string {
	v.rw.Write([]byte(line))
	v.rw.Write([]byte("\n"))
	v.rw.Flush()
	message, _ := v.rw.ReadString('\n')
	log.Printf("sending %s, received: %s", strings.Replace(line, "\n", "", -1), message)
	return message
}

func run(c string, args ...string) *exec.Cmd {
	log.Printf("running %s %s", c, strings.Join(args, " "))
	cmd := exec.Command(c, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}

func RandomString(len int) string {
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		bytes[i] = byte(65 + rand.Intn(25)) //A=65 and Z = 65+25
	}
	return string(bytes)
}

func exists(f string) bool {
	if _, err := os.Stat(f); !os.IsNotExist(err) {
		return true
	}
	return false

}
func waitForFile(f string) {
	for {
		if exists(f) {
			return
		}
		log.Printf("waiting for %s", f)
		time.Sleep(1 * time.Second)
	}
}

func readPortFile(fn string) int {
	content, err := ioutil.ReadFile(fn)
	if err != nil {
		log.Fatal(err)
	}
	n, err := strconv.Atoi(strings.Replace(string(content), "\n", "", -1))
	if err != nil {
		log.Fatal(err)
	}
	return n
}

func NewVowpalInstance(modelPath string) *vowpal {
	rand.Seed(time.Now().UTC().UnixNano())

	fn := path.Join(os.TempDir(), fmt.Sprintf("juun.%s.vw.port", RandomString(16)))

	log.Printf("starting vw with port file %s", fn)
	args := []string{
		"--random_seed",
		"123",
		"--quiet",
		"-b",
		"18",
		"--bootstrap",
		"2",
		"--port",
		"0",
		"--port_file",
		fn,
		"-q",
		"ci",
		"--no_stdin",
		"--foreground",
		"--num_children",
		"1",
		"--loss_function",
		"logistic",
		"--link",
		"logistic",
		"--ftrl",
		"--save_resume",
		"-f",
		modelPath,
	}

	if exists(modelPath) {
		args = append(args, "-i", modelPath)
	}

	vwCMD := run("/usr/local/bin/vw", args...)
	vwCMD.Stdout = os.Stderr
	vwCMD.Stderr = os.Stderr
	if err := vwCMD.Start(); err != nil {
		fmt.Println("An error occured: ", err)
	}

	waitForFile(fn)
	port := readPortFile(fn)

	var conn net.Conn
	var err error
	for {
		conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			break
		}
		log.Printf("trying to connect to %d, err: %s", port, err.Error())
		time.Sleep(1 * time.Second)
	}

	return &vowpal{fn: fn, conn: conn, cmd: vwCMD, rw: bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), modelPath: modelPath}
}

func (v *vowpal) getVowpalScore(features string) float32 {
	s := strings.Replace(v.SendReceive(features), "\n", "", -1)
	splitted := strings.Split(s, " ")
	f, err := strconv.ParseFloat(splitted[2], 32)
	if err != nil {
		log.Printf("err: %s", err)
	}
	return float32(f)
}

type bandit struct {
	*vowpal
	predictions map[int]*prediction // item id -> last prediction
}

func NewBandit(modelPath string) *bandit {
	return &bandit{vowpal: NewVowpalInstance(modelPath), predictions: map[int]*prediction{}}
}

type item struct {
	id       int
	features string
}

type prediction struct {
	items map[int]*item
	ts    int64
}

func (v *bandit) Predict(items ...*item) map[int]float32 {
	out := map[int]float32{}

	prediction := &prediction{
		ts:    time.Now().Unix(),
		items: map[int]*item{},
	}

	for _, item := range items {
		out[item.id] = v.vowpal.getVowpalScore(item.features)
		prediction.items[item.id] = item
		v.predictions[item.id] = prediction
	}

	return out
}

func (v *bandit) Click(clicked int) {
	pred, ok := v.predictions[clicked]
	if !ok {
		return
	}

	for _, item := range pred.items {
		label := -1
		if item.id == clicked {
			label = 1
		}
		v.vowpal.SendReceive(fmt.Sprintf("%d %s", label, item.features))
	}

	v.Expire()
}

func (v *bandit) Expire() {
	// expire the old ones
	// FIXME: train negative?
	now := time.Now().Unix()
	for k, p := range v.predictions {
		if now-p.ts > 60 {
			delete(v.predictions, k)
		}
	}
}
