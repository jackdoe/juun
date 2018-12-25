package main

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type HistoryLine struct {
	Line      string
	TimeStamp int64
	Count     uint32
	Id        int
}

type History struct {
	Lines       []*HistoryLine
	index       map[string]int
	inverted    *InvertedIndex
	perTerminal map[int]*Terminal
	lock        sync.Mutex
	vw          *vowpal
}

func NewHistory() *History {
	return &History{
		Lines:       []*HistoryLine{}, // ordered list of commands
		index:       map[string]int{}, // XXX: dont store the strings twice
		perTerminal: map[int]*Terminal{},
		inverted: &InvertedIndex{
			Postings:  map[string][]uint64{},
			TotalDocs: 0,
		},
	}
}

func (h *History) selfReindex() {
	log.Printf("starting reindexing")
	h.inverted = &InvertedIndex{
		Postings:  map[string][]uint64{},
		TotalDocs: 0,
	}
	for id, v := range h.Lines {
		h.addLineToInvertedIndex(v)
		h.index[v.Line] = id
	}
	log.Printf("reindexing done, %d items", len(h.index))

}

func (h *History) addLineToInvertedIndex(v *HistoryLine) {
	for _, s := range tokenize(v.Line) {
		h.inverted.add(v.Id, fmt.Sprintf("t_%s", s))
		for _, e := range edge(s) {
			h.inverted.add(v.Id, fmt.Sprintf("e_%s", e))
		}
	}
	h.inverted.TotalDocs++
}

func (h *History) deletePID(pid int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	delete(h.perTerminal, pid)
}

var SPLIT_REGEXP = regexp.MustCompile("[~&@%/_,\\.-]+")

func tokenize(s string) []string {
	trimmed := strings.Replace(s, "\n", " ", -1)
	seen := map[string]bool{}
	splitted := strings.Split(trimmed, " ")
	for _, sp := range splitted {
		seen[sp] = true
		for _, more := range SPLIT_REGEXP.Split(sp, -1) {
			seen[more] = true
		}
	}
	out := []string{}
	for k, _ := range seen {
		if len(k) > 0 {
			out = append(out, k)
		}
	}

	return out
}

func edge(text string) []string {
	out := []string{}
	for i := 0; i < len(text); i++ {
		out = append(out, text[:i+1])
	}
	return out
}

func (h *History) add(line string, pid int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	t, ok := h.perTerminal[pid]
	if !ok {
		// when new terminal starts we want its GlobalID to point just before the command was added
		// otherwise it points to the first command and we have to up()up() twice to go to the global history
		id := len(h.Lines)

		t = &Terminal{
			Commands:        []int{},
			Cursor:          0,
			GlobalIdOnStart: id,
			CommandsSet:     map[int]bool{},
		}
		h.perTerminal[pid] = t
	}

	now := time.Now().UnixNano()
	id, ok := h.index[line]
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
		h.index[line] = v.Id
		h.addLineToInvertedIndex(v)
	}

	t.add(id)
}

func (h *History) gotoend(pid int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	t, ok := h.perTerminal[pid]
	if !ok {
		return
	}
	t.end()
}
func (h *History) up(pid int, buf string) string {
	return h.move(true, pid, buf)
}

func (h *History) down(pid int, buf string) string {
	return h.move(false, pid, buf)
}

func (h *History) move(goUP bool, pid int, buf string) string {
	h.lock.Lock()
	defer h.lock.Unlock()

	t, ok := h.perTerminal[pid]
	if !ok {
		id := len(h.Lines)
		t = &Terminal{
			Commands:        []int{},
			Cursor:          0,
			GlobalIdOnStart: id,
			CommandsSet:     map[int]bool{},
		}
		h.perTerminal[pid] = t
	}

	if goUP && t.isAtEnd() {
		t.CurrentBufferBeforeMove = buf
	}

	var can bool
	var id int
	if goUP {
		id, can = t.up()
	} else {
		id, can = t.down()

		if !can {
			return t.CurrentBufferBeforeMove
		}
	}

	if len(h.Lines) == 0 {
		return ""
	}

	return h.Lines[id].Line
}

type scored struct {
	docId int32
	score float32
}

type ByScore []scored

func (s ByScore) Len() int {
	return len(s)
}
func (s ByScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByScore) Less(i, j int) bool {
	return s[j].score < s[i].score
}

func (h *History) search(text string, pid int) string {
	h.lock.Lock()
	defer h.lock.Unlock()

	text = strings.Trim(text, " ")
	if len(text) == 0 {
		return ""
	}

	terms := []Query{}
	for _, s := range tokenize(text) {
		terms = append(terms, h.inverted.term("e", s))
	}

	query := NewBoolOrQuery(terms)
	score := []scored{}
	terminal, hasTerminal := h.perTerminal[pid]

	now := time.Now().Unix()
	maxScore := float32(0)
	for query.Next() != NO_MORE {
		id := query.GetDocId()
		line := h.Lines[id]

		tfidf := query.Score()

		ts := line.TimeStamp / 1000000000
		timeScore := float32(-math.Log10(1 + float64(now-ts))) // -log(1+secondsDiff)

		countScore := float32(math.Log1p(float64(line.Count)))
		terminalScore := float32(0)
		if hasTerminal {
			_, hasCommandInHistory := terminal.CommandsSet[int(id)]
			if hasCommandInHistory {
				terminalScore = 100
			}
		}

		s := tfidf + timeScore + terminalScore
		if s > maxScore {
			log.Printf("tfidf: %f timeScore: %f terminalScore:%f countScore:%f, age: %ds - %s", tfidf, timeScore, terminalScore, countScore, now-ts, line.Line)
			maxScore = s
		}
		score = append(score, scored{query.GetDocId(), s})
	}
	sort.Sort(ByScore(score))
	// take the top 10 and sort them using vowpal wabbit's bootstrap
	if len(score) > 0 {
		return h.Lines[score[0].docId].Line
	}
	return ""
}
