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
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type vowpal struct {
	conn      net.Conn
	cmd       *exec.Cmd
	rw        *bufio.ReadWriter
	fn        string
	modelPath string
	sync.Mutex
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
		return fmt.Sprintf("%s:%f", f.feature, f.value)
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
}

func NewFeatureSet(nss ...*namespace) *featureSet {
	return &featureSet{
		namespaces: nss,
	}
}

func (fs *featureSet) addNamespaces(nss ...*namespace) {
	fs.namespaces = append(fs.namespaces, nss...)
}

func (fs *featureSet) add(other *featureSet) {
	fs.namespaces = append(fs.namespaces, other.namespaces...)
}

func (fs *featureSet) toVW() string {
	var sb strings.Builder
	for _, f := range fs.namespaces {
		sb.WriteString(f.toVW())
		sb.WriteString(" ")
	}

	return sb.String()
}
func (v *vowpal) Save() {
	v.Lock()
	v.rw.Write([]byte(fmt.Sprintf("save_%s", v.modelPath)))
	v.rw.Write([]byte("\n"))
	v.rw.Flush()
	v.Unlock()

	waitForFile(v.modelPath)
}

func (v *vowpal) SendReceive(line string) string {
	v.Lock()
	defer v.Unlock()

	v.rw.Write([]byte(line))
	v.rw.Write([]byte("\n"))
	v.rw.Flush()
	log.Printf("sending %s", strings.Replace(line, "\n", "", -1))
	message, _ := v.rw.ReadString('\n')
	log.Printf("received %s", message)
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
			log.Printf("found %s", f)
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

	binary, lookErr := exec.LookPath("vw")
	if lookErr != nil {
		return nil
	}

	fn := path.Join(os.TempDir(), fmt.Sprintf("juun.%s.vw.port", RandomString(16)))

	log.Printf("starting vw with port file %s", fn)
	args := []string{
		"--quiet",
		"--port",
		"0",
		"--port_file",
		fn,
		"--no_stdin",
		"--foreground",
		"--num_children",
		"1",
		"--loss_function",
		"logistic",

		"--save_resume",
		"-f",
		modelPath,
	}

	if exists(modelPath) {
		args = append(args, "-i", modelPath)
	} else {
		args = append(args,
			"--link",
			"logistic",
			"-b",
			"18",

			"-q",
			"ci",
			"--bootstrap",
			"2",
			"--ftrl")
	}
	vwCMD := run(binary, args...)
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

type banditScore struct {
	score float32
	item  *item
}
type ByBanditScore []banditScore

func (s ByBanditScore) Len() int {
	return len(s)
}
func (s ByBanditScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByBanditScore) Less(i, j int) bool {
	return s[j].score < s[i].score
}

func (v *bandit) Predict(limit int, items ...*item) map[int]float32 {
	scores := []banditScore{}
	prediction := &prediction{
		ts:    time.Now().Unix(),
		items: map[int]*item{},
	}

	for _, item := range items {
		scores = append(scores, banditScore{
			score: v.vowpal.getVowpalScore(item.features),
			item:  item,
		})
	}

	sort.Sort(ByBanditScore(scores))
	if limit > len(scores) {
		limit = len(scores)
	}
	out := map[int]float32{}
	for i := 0; i < limit; i++ {
		s := scores[i]
		id := s.item.id
		out[id] = s.score
		prediction.items[id] = s.item
		v.predictions[id] = prediction
	}

	for _, item := range items {
		_, ok := out[item.id]
		if !ok {
			out[item.id] = 0
		}
	}

	return out
}

func (v *bandit) Click(clicked int) {
	pred, ok := v.predictions[clicked]
	if ok {
		for _, item := range pred.items {
			label := -1
			if item.id == clicked {
				label = 1
			}
			v.vowpal.SendReceive(fmt.Sprintf("%d %s", label, item.features))
			delete(v.predictions, item.id)
		}

	}
	v.Expire()
}

func (v *bandit) Expire() {
	// expire the old ones
	// FIXME: train negative?
	now := time.Now().Unix()
	for k, p := range v.predictions {
		// due to the nature of auto-complete search, it does not make sense to train with
		// items that were probably not even seen
		// all items in the last 10 seconds that were not clicked are treated as bad
		// everything else is ignored
		if now-p.ts < 10 {
			for _, item := range p.items {
				v.vowpal.SendReceive(fmt.Sprintf("-1 %s", item.features))
			}
		}
		delete(v.predictions, k)
	}
}
