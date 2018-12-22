package main

import (
	//	"log"
	"math"
)

const (
	NO_MORE   = int32(math.MaxInt32)
	NOT_READY = int32(-1)
)

type Query interface {
	advance(int32) int32
	Next() int32
	GetDocId() int32
	Score() float32
}

type QueryBase struct {
	docId int32
}

func (q *QueryBase) GetDocId() int32 {
	return q.docId
}

type Term struct {
	cursor    int32
	postings  []uint64
	term      string
	totalDocs int32
	QueryBase
}

func (t *Term) Score() float32 {
	//	log.Printf("%d tf totalDocs: %d docsWithTerm: %d", t.getAt(t.cursor)&0xFFFF, t.totalDocs, len(t.postings))
	return float32(float64(t.getAt(t.cursor)&0xFFFF) * math.Log1p(float64(len(t.postings))/float64(t.totalDocs)))
}

func NewTerm(term string, totalDocs int32, postings []uint64) *Term {
	return &Term{
		cursor:    0,
		term:      term,
		QueryBase: QueryBase{NOT_READY},
		postings:  postings,
		totalDocs: totalDocs,
	}
}

func (t *Term) getAt(idx int32) uint64 {
	return t.postings[idx]
}

func (t *Term) advance(target int32) int32 {
	if t.docId == NO_MORE || t.docId == target || target == NO_MORE {
		t.docId = target
		return t.docId
	}
	start := t.cursor
	end := int32(len(t.postings))
	for start < end {
		mid := start + ((end - start) / 2)
		current := int32(t.getAt(mid) >> 16)
		if current == target {
			t.cursor = mid
			t.docId = target
			return t.GetDocId()
		}

		if current < target {
			start = mid + 1
		} else {
			end = mid
		}
	}

	return t.move(start)
}

func (t *Term) move(to int32) int32 {
	t.cursor = to
	if t.cursor >= int32(len(t.postings)) {
		t.docId = NO_MORE
	} else {
		t.docId = int32(t.getAt(t.cursor) >> 16)
	}
	return t.docId
}

func (t *Term) Next() int32 {
	if t.docId != NOT_READY {
		t.cursor++
	}
	return t.move(t.cursor)
}

type BoolQueryBase struct {
	queries []Query
}

func (q *BoolQueryBase) AddSubQuery(sub Query) {
	q.queries = append(q.queries, sub)
}

type BoolOrQuery struct {
	BoolQueryBase
	QueryBase
}

func NewBoolOrQuery(queries []Query) *BoolOrQuery {
	return &BoolOrQuery{
		BoolQueryBase: BoolQueryBase{queries},
		QueryBase:     QueryBase{NOT_READY},
	}
}

func (q *BoolOrQuery) Score() float32 {
	total := float32(0)
	for i := 0; i < len(q.queries); i++ {
		if q.queries[i].GetDocId() == q.GetDocId() {
			total += q.queries[i].Score()
		}
	}
	return total
}

func (q *BoolOrQuery) advance(target int32) int32 {
	new_doc := NO_MORE
	for _, sub_query := range q.queries {
		cur_doc := sub_query.GetDocId()
		if cur_doc < target {
			cur_doc = sub_query.advance(target)
		}

		if cur_doc < new_doc {
			new_doc = cur_doc
		}
	}
	q.docId = new_doc
	return q.docId
}

func (q *BoolOrQuery) Next() int32 {
	new_doc := NO_MORE
	for _, sub_query := range q.queries {
		cur_doc := sub_query.GetDocId()
		if cur_doc == q.docId {
			cur_doc = sub_query.Next()
		}

		if cur_doc < new_doc {
			new_doc = cur_doc
		}
	}
	q.docId = new_doc
	return new_doc
}

type BoolAndQuery struct {
	BoolQueryBase
	QueryBase
}

func NewBoolAndQuery(queries []Query) *BoolAndQuery {
	return &BoolAndQuery{
		BoolQueryBase: BoolQueryBase{queries},
		QueryBase:     QueryBase{NOT_READY},
	}
}

func (q *BoolAndQuery) Score() float32 {
	total := float32(0)
	for i := 0; i < len(q.queries); i++ {
		total += q.queries[i].Score()
	}
	return total
}

func (q *BoolAndQuery) nextAndedDoc(target int32) int32 {
	// initial iteration skips queries[0]
	for i := 1; i < len(q.queries); i++ {
		sub_query := q.queries[i]

		if sub_query.GetDocId() < target {
			sub_query.advance(target)
		}

		if sub_query.GetDocId() == target {
			continue
		}

		target = q.queries[0].advance(sub_query.GetDocId())
		i = 0 //restart the loop from the first query
	}
	q.docId = target
	return q.docId
}

func (q *BoolAndQuery) advance(target int32) int32 {
	if len(q.queries) == 0 {
		q.docId = NO_MORE
		return NO_MORE
	}

	return q.nextAndedDoc(q.queries[0].advance(target))
}

func (q *BoolAndQuery) Next() int32 {
	if len(q.queries) == 0 {
		q.docId = NO_MORE
		return NO_MORE
	}

	return q.nextAndedDoc(q.queries[0].Next())
}
