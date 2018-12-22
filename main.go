package main

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type HistoryLine struct {
	Line      string
	TimeStamp int64
	Count     uint64
	Id        int
}

type History struct {
	Lines       []*HistoryLine
	Index       map[string]int
	PerTerminal map[int]*Terminal
	sync.Mutex
}

type Terminal struct {
	Commands        []int
	Cursor          int
	GlobalIdOnStart int
	GlobalId        int
	globalMode      bool
}

func (t *Terminal) currentCommandId() int {
	if !t.globalMode {
		return t.Commands[t.Cursor]
	}
	return t.GlobalId
}

func (t *Terminal) up() {
	old := t.Cursor
	if t.Cursor > 0 {
		t.Cursor--
	} else {
		if !t.globalMode {
			//			log.Printf("enabling global mode")
			t.globalMode = true
		}
		if t.GlobalId > 0 {
			t.GlobalId--
		}
	}
	//	log.Printf("DOWN from %d to %d", old, t.Cursor)
}

func (t *Terminal) down() {
	old := t.Cursor

	if t.globalMode {
		if t.GlobalId >= t.GlobalIdOnStart {
			t.globalMode = false
			//			log.Printf("disabling global mode")
		} else {
			t.GlobalId++
		}
	} else {
		if t.Cursor < len(t.Commands)-1 {
			t.Cursor++
		}
	}
	//	log.Printf("UP from %d to %d", old, t.Cursor)
}

func NewHistory() *History {
	return &History{
		Lines:       []*HistoryLine{}, // ordered list of commands
		Index:       map[string]int{}, // XXX: dont store the strings twice
		PerTerminal: map[int]*Terminal{},
	}
}

func (h *History) deletePID(pid int) {
	delete(h.PerTerminal, pid)
}

func (h *History) add(line string, pid int) {
	h.Lock()
	defer h.Unlock()
	now := time.Now().UnixNano()
	id, ok := h.Index[line]
	if ok {
		v := h.Lines[id]
		v.Count++
		v.TimeStamp = now
	} else {
		id = len(h.Lines)
		v := &HistoryLine{
			Line:      line,
			TimeStamp: now,
			Count:     1,
			Id:        id,
		}
		h.Lines = append(h.Lines, v)
		h.Index[line] = v.Id
	}

	t, ok := h.PerTerminal[pid]
	if !ok {
		t = &Terminal{
			Commands:        []int{},
			Cursor:          0,
			GlobalId:        id,
			GlobalIdOnStart: id,
		}
		h.PerTerminal[pid] = t
	}
	t.Cursor = len(t.Commands)
	t.Commands = append(t.Commands, id)
}

func (h *History) move(goup bool, pid int) string {
	t, ok := h.PerTerminal[pid]
	if !ok {
		return ""
	}
	if goup {
		t.up()
	} else {
		t.down()
	}
	return h.Lines[t.currentCommandId()].Line
}

func (h *History) search(query string, pid int) string {
	// XXX: poc, FIXME: 3gram, tfidf, frequency, vw etc ete
	//	log.Printf("searching for %s", query)
	t, ok := h.PerTerminal[pid]
	if ok {
		for i := len(t.Commands) - 1; i >= 0; i-- {
			c := h.Lines[t.Commands[i]]
			if strings.Contains(c.Line, query) {
				return c.Line
			}
		}
	}

	for i := len(h.Lines) - 1; i >= 0; i-- {
		c := h.Lines[i]
		if strings.Contains(c.Line, query) {
			return c.Line
		}
	}
	return ""
}

func intOrZero(s string) int {
	pid, _ := strconv.Atoi(s)
	return pid
}

func main() {
	history := NewHistory()
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	histfile := path.Join(usr.HomeDir, ".juun.json")
	dat, err := ioutil.ReadFile(histfile)
	if err == nil {
		err = json.Unmarshal(dat, &history)
		if err == nil {
			history = NewHistory()
		}
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		d1, err := json.Marshal(history)
		if err == nil {
			ioutil.WriteFile(histfile, d1, 0600)
		}
		os.Exit(0)
	}()

	r := gin.Default()

	r.GET("/down/:pid", func(c *gin.Context) {
		pid := intOrZero(c.Param("pid"))
		c.String(http.StatusOK, "%s\n", history.move(false, pid))
	})

	r.GET("/up/:pid", func(c *gin.Context) {
		pid := intOrZero(c.Param("pid"))
		c.String(http.StatusOK, "%s\n", history.move(true, pid))
	})

	r.GET("/delete/:pid", func(c *gin.Context) {
		pid := intOrZero(c.Param("pid"))
		history.deletePID(pid)
		c.String(http.StatusOK, "ok: %d\n", pid)
	})

	r.GET("/add/:pid", func(c *gin.Context) {
		pid := intOrZero(c.Param("pid"))
		line, err := c.GetRawData()
		if err != nil {
			c.String(http.StatusBadRequest, "err: %s\n", err.Error())
			return
		}
		if len(line) > 0 {
			history.add(string(line), pid)
		}
		//		log.Printf("adding %s", line)
		c.String(http.StatusOK, "ok: %s\n", line)
	})

	r.GET("/search/:pid", func(c *gin.Context) {
		pid := intOrZero(c.Param("pid"))
		line, err := c.GetRawData()
		if err != nil {
			c.String(http.StatusBadRequest, "err: %s\n", err.Error())
			return
		}
		c.String(http.StatusOK, history.search(string(line), pid))
	})
	r.GET("/dump", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"history": history,
		})
	})
	r.Run()
}
