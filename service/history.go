package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	iq "github.com/jackdoe/go-query"
	"github.com/jackdoe/go-query/util/analyzer"
	"github.com/jackdoe/go-query/util/index"
	"github.com/jackdoe/go-query/util/tokenize"
	. "github.com/jackdoe/juun/common"
	. "github.com/jackdoe/juun/vw"
	log "github.com/sirupsen/logrus"
)

func toDocuments(in []*HistoryLine) []index.Document {
	out := make([]index.Document, len(in))
	for i, d := range in {
		out[i] = index.Document(d)
	}
	return out
}

type History struct {
	Lines       []*HistoryLine
	lookup      map[string]int
	PerTerminal map[int]*Terminal
	idx         *index.MemOnlyIndex
	lock        sync.Mutex
	vw          *Bandit
}

func NewHistory() *History {
	indexTokenizer := []tokenize.Tokenizer{
		tokenize.NewWhitespace(),
		tokenize.NewLeftEdge(1), // left edge ngram indexing for prefix matches
		tokenize.NewUnique(),
	}

	searchTokenizer := []tokenize.Tokenizer{
		tokenize.NewWhitespace(),
		tokenize.NewUnique(),
	}

	autocomplete := analyzer.NewAnalyzer(
		index.DefaultNormalizer,
		searchTokenizer,
		indexTokenizer,
	)
	m := index.NewMemOnlyIndex(map[string]*analyzer.Analyzer{
		"line": autocomplete,
	})

	return &History{
		Lines:       []*HistoryLine{}, // ordered list of commands
		lookup:      map[string]int{},
		idx:         m,
		PerTerminal: map[int]*Terminal{},
	}
}

func (h *History) selfReindex() {
	log.Infof("starting reindexing")
	h.lookup = map[string]int{}
	for id, v := range h.Lines {
		h.lookup[v.Line] = id
	}
	h.idx.Index(toDocuments(h.Lines)...)
	log.Infof("reindexing done, %d items", len(h.Lines))

}

func (h *History) deletePID(pid int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	delete(h.PerTerminal, pid)
}

func (h *History) add(line string, pid int, env map[string]string) {
	h.lock.Lock()
	defer h.lock.Unlock()

	t := h.getTerminal(pid)

	now := time.Now().UnixNano()
	id, ok := h.lookup[line]
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
		h.lookup[line] = v.Id
		h.idx.Index(index.Document(v))
	}

	if h.vw != nil {
		h.vw.Click(id)
		h.like(h.Lines[id], env)
	}

	t.CurrentBufferBeforeMove = ""
	t.add(id)
}

func (h *History) gotoend(pid int) {
	h.lock.Lock()
	defer h.lock.Unlock()

	t, ok := h.PerTerminal[pid]
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

func (h *History) getTerminal(pid int) *Terminal {
	t, ok := h.PerTerminal[pid]
	if !ok {
		id := len(h.Lines)
		t = NewTerminal(id)
		h.PerTerminal[pid] = t

	}
	return t
}
func (h *History) move(goUP bool, pid int, buf string) string {
	h.lock.Lock()
	defer h.lock.Unlock()

	t := h.getTerminal(pid)

	var can bool
	var id int

	if goUP && t.isAtEnd() {
		t.CurrentBufferBeforeMove = buf
	}

	if goUP {
		id, _ = t.up()
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
	id            int
	score         float32
	tfidf         float32
	countScore    float32
	timeScore     float32
	terminalScore float32
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

const scoreOnTerminal = float32(10)

func (h *History) search(text string, pid int, env map[string]string) []*HistoryLine {
	h.lock.Lock()
	defer h.lock.Unlock()

	text = strings.Trim(text, " ")
	if len(text) == 0 {
		return []*HistoryLine{}
	}
	query := iq.And(h.idx.Terms("line", text)...)

	score := []scored{}
	terminal, hasTerminal := h.PerTerminal[pid]
	now := time.Now().Unix()

	h.idx.Foreach(query, func(did int32, tfidf float32, doc index.Document) {
		line := doc.(*HistoryLine)
		ts := line.TimeStamp / 1000000000
		timeScore := float32(-math.Sqrt(1 + float64(now-ts))) // -log(1+secondsDiff)

		countScore := float32(math.Sqrt(float64(line.Count)))
		terminalScore := float32(0)
		if hasTerminal {
			_, hasCommandInHistory := terminal.CommandsSet[line.Id]
			if hasCommandInHistory {
				terminalScore = scoreOnTerminal
			}
		}

		total := tfidf + (5 * timeScore) + terminalScore + countScore
		score = append(score, scored{id: line.Id, score: total, tfidf: tfidf, timeScore: timeScore, terminalScore: terminalScore, countScore: countScore})

	})

	sort.Sort(ByScore(score))

	if h.vw != nil {
		// take the top 5 and sort them using vowpal wabbit's bootstrap
		topN := 2
		if topN > len(score) {
			topN = len(score)
		}

		ctx := UserContext(text, GetOrDefault(env, "cwd", ""))
		vwi := []*Item{}
		for i := 0; i < topN; i++ {
			s := score[i]
			line := h.Lines[s.id]

			f := line.Featurize()
			f.Add(ctx)
			f.AddNamespaces(
				NewNamespace("i_score",
					NewFeature("tfidf", s.tfidf),
					NewFeature("timeScore", s.timeScore),
					NewFeature("countScore", s.countScore),
					NewFeature(fmt.Sprintf("terminalScore=%d", int(s.terminalScore)), 0)))

			vwi = append(vwi, NewItem(line.Id, f.ToVW()))
			log.Debugf("before VW: tfidf: %f timeScore: %f terminalScore:%f countScore:%f line:%s", s.tfidf, s.timeScore, s.terminalScore, s.countScore, line.Line)
		}

		prediction := h.vw.Predict(1, vwi...)
		sort.Slice(score, func(i, j int) bool { return prediction[int(score[j].id)] < prediction[int(score[i].id)] })
	}

	// pick the first one
	out := []*HistoryLine{}
	if len(score) > 0 {
		for _, s := range score {
			line := h.Lines[s.id]
			out = append(out, line)
		}
	}
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}

func (h *History) like(line *HistoryLine, env map[string]string) {
	if h.vw == nil {
		return
	}

	ctx := UserContext("", GetOrDefault(env, "cwd", ""))
	f := line.Featurize()
	f.Add(ctx)
	f.AddNamespaces(NewNamespace("i_score", NewFeature(fmt.Sprintf("terminalScore=%d", int(scoreOnTerminal)), 0)))
	h.vw.SendReceive(fmt.Sprintf("1 10 %s", f.ToVW())) // add weight of 10 on the clicked one
}
