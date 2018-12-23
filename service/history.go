package main

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"sync"
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
	Inverted    *InvertedIndex
	perTerminal map[int]*Terminal
	lock        sync.Mutex
}

func NewHistory() *History {
	return &History{
		Lines:       []*HistoryLine{}, // ordered list of commands
		Index:       map[string]int{}, // XXX: dont store the strings twice
		perTerminal: map[int]*Terminal{},
		Inverted: &InvertedIndex{
			Postings:  map[string][]uint64{},
			TotalDocs: 0,
		},
	}
}

func (h *History) deletePID(pid int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	delete(h.perTerminal, pid)
}

func tokenize(s string) []string {
	trimmed := strings.Replace(s, "\n", " ", -1)
	return strings.Split(trimmed, " ")
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
		for _, s := range tokenize(line) {
			h.Inverted.add(id, fmt.Sprintf("t_%s", s))
			for _, e := range edge(s) {
				h.Inverted.add(id, fmt.Sprintf("e_%s", e))
			}
		}
		h.Inverted.TotalDocs++
	}

	t, ok := h.perTerminal[pid]
	if !ok {
		t = &Terminal{
			Commands:        []int{},
			Cursor:          0,
			GlobalId:        id,
			GlobalIdOnStart: id,
			CommandsSet:     map[int]bool{},
		}
		h.perTerminal[pid] = t
	}

	t.Commands = append(t.Commands, id)
	t.Cursor = len(t.Commands)
	t.CommandsSet[id] = true
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

func (h *History) move(goUP bool, pid int) string {
	h.lock.Lock()
	defer h.lock.Unlock()

	if len(h.Lines) == 0 {
		return ""
	}

	t, ok := h.perTerminal[pid]
	if !ok {
		id := len(h.Lines) - 1
		t = &Terminal{
			Commands:        []int{},
			Cursor:          0,
			GlobalId:        id,
			GlobalIdOnStart: id,
			CommandsSet:     map[int]bool{},
		}
		h.perTerminal[pid] = t
	}
	hasDownToGo := false
	if goUP {
		hasDownToGo = true
		t.up()
	} else {
		hasDownToGo = t.down()
	}
	if !hasDownToGo {
		return ""
	}
	return h.Lines[t.currentCommandId()].Line
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
		terms = append(terms, h.Inverted.term("e", s))
	}

	query := NewBoolOrQuery(terms)
	score := []scored{}
	terminal, hasTerminal := h.perTerminal[pid]

	now := time.Now().UnixNano()
	for query.Next() != NO_MORE {
		id := query.GetDocId()
		line := h.Lines[id]

		tfidf := query.Score()

		timeScore := float32(-math.Log10(1 + float64(now-line.TimeStamp)))

		terminalScore := float32(0)
		if hasTerminal {
			_, hasCommandInHistory := terminal.CommandsSet[int(id)]
			if hasCommandInHistory {
				terminalScore = 100
			}
		}

		log.Printf("tfidf: %f timeScore: %f terminalScore:%f %s", tfidf, timeScore, terminalScore, line.Line)
		s := tfidf + timeScore + terminalScore
		score = append(score, scored{query.GetDocId(), s})
	}
	sort.Sort(ByScore(score))

	if len(score) > 0 {
		return h.Lines[score[0].docId].Line
	}
	return ""
}
