package main

import (
	"fmt"
	. "github.com/jackdoe/juun/common"
	. "github.com/jackdoe/juun/vw"
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

func (l *HistoryLine) featurize() *FeatureSet {
	features := []*Feature{}
	for _, s := range strings.Split(l.Line, " ") {
		features = append(features, NewFeature(s, 0))
	}
	text := NewNamespace("i_text", features...)

	count := NewNamespace("i_count", NewFeature("count", float32(math.Log(float64(1)+float64(l.Count)))))
	t := timeToNamespace("i_time", time.Unix(l.TimeStamp/1000000000, 0))
	id := NewNamespace("i_id", NewFeature(fmt.Sprintf("id=%d", l.Id), float32(0)))

	return NewFeatureSet(id, text, count, t)
}

func timeToNamespace(ns string, now time.Time) *Namespace {
	features := []*Feature{}
	hr, _, _ := now.Clock()

	features = append(features, NewFeature(fmt.Sprintf("year=%d", now.Year()), 0))
	features = append(features, NewFeature(fmt.Sprintf("day=%d", now.Day()), 0))
	features = append(features, NewFeature(fmt.Sprintf("month=%d", now.Month()), 0))
	features = append(features, NewFeature(fmt.Sprintf("hour=%d", hr), 0))

	return NewNamespace(ns, features...)
}

func userContext(query string, cwd string) *FeatureSet {
	features := []*Feature{}
	for _, s := range strings.Split(query, " ") {
		if len(s) > 0 {
			features = append(features, NewFeature(s, 0))
		}
	}
	qns := NewNamespace("c_query", features...)

	fs := NewFeatureSet(timeToNamespace("c_user_time", time.Now()), qns)
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

type History struct {
	Lines       []*HistoryLine
	index       map[string]int
	inverted    *InvertedIndex
	perTerminal map[int]*Terminal
	lock        sync.Mutex
	vw          *Bandit
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

func (h *History) add(line string, pid int, env map[string]string) {
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

	if h.vw != nil {
		h.vw.Click(id)
		h.like(h.Lines[id], env)
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

const scoreOnTerminal = float32(100)

func (h *History) search(text string, pid int, env map[string]string) string {
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

	query := NewBoolAndQuery(terms)
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
				terminalScore = scoreOnTerminal
			}
		}

		total := tfidf + timeScore + terminalScore + countScore
		if total > maxScore {
			log.Printf("total: %f, tfidf: %f timeScore: %f terminalScore:%f countScore:%f, age: %ds - %s", total, tfidf, timeScore, terminalScore, countScore, now-ts, line.Line)
			maxScore = total
		}
		score = append(score, scored{id: line.Id, score: total, tfidf: tfidf, timeScore: timeScore, terminalScore: terminalScore, countScore: countScore})
	}
	sort.Sort(ByScore(score))

	if h.vw != nil {
		// take the top 5 and sort them using vowpal wabbit's bootstrap
		topN := 5
		if topN > len(score) {
			topN = len(score)
		}

		ctx := userContext(text, GetOrDefault(env, "cwd", ""))
		vwi := []*Item{}
		for i := 0; i < topN; i++ {
			s := score[i]
			line := h.Lines[s.id]

			f := line.featurize()
			f.Add(ctx)
			f.AddNamespaces(
				NewNamespace("i_score",
					NewFeature("tfidf", s.tfidf),
					NewFeature("timeScore", s.timeScore),
					NewFeature("countScore", s.countScore),
					NewFeature(fmt.Sprintf("terminalScore=%d", int(s.terminalScore)), 0)))

			vwi = append(vwi, NewItem(line.Id, f.ToVW()))
		}

		prediction := h.vw.Predict(1, vwi...)
		sort.Slice(score, func(i, j int) bool { return prediction[int(score[j].id)] < prediction[int(score[i].id)] })
	}

	// pick the first one
	if len(score) > 0 {
		return h.Lines[score[0].id].Line
	}

	return ""
}

func (h *History) like(line *HistoryLine, env map[string]string) {
	if h.vw == nil {
		return
	}

	ctx := userContext("", GetOrDefault(env, "cwd", ""))
	f := line.featurize()
	f.Add(ctx)
	f.AddNamespaces(NewNamespace("i_score", NewFeature(fmt.Sprintf("terminalScore=%d", int(scoreOnTerminal)), 0)))
	h.vw.SendReceive(fmt.Sprintf("1 %s", f.ToVW()))
}
