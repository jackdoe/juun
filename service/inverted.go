package main

import (
	"fmt"
)

type InvertedIndex struct {
	Postings  map[string][]uint64
	TotalDocs int32
}

func (x *InvertedIndex) encode(docId int, freq int) uint64 {
	return uint64(docId)<<uint64(16) | uint64(freq)
}

func (x *InvertedIndex) decode(p uint64) (int, int) {
	docId := p >> 16
	freq := p & 0xFFFF
	return int(docId), int(freq)
}
func (x *InvertedIndex) term(field string, term string) Query {
	t := fmt.Sprintf("%s_%s", field, term)
	postings, ok := x.Postings[t]
	if ok {
		//		log.Printf("%s - %#v", t, postings)
		return NewTerm(t, x.TotalDocs, postings)
	}
	//	log.Printf("%s - missing", t)
	return NewTerm(t, x.TotalDocs, []uint64{})
}

func (x *InvertedIndex) add(docId int, term string) {
	postings, ok := x.Postings[term]
	if ok {
		last := postings[len(postings)-1]
		lastDocId, lastFreq := x.decode(last)
		if lastDocId == docId {
			postings[len(postings)-1] = x.encode(docId, lastFreq+1)
		} else {
			x.Postings[term] = append(x.Postings[term], x.encode(docId, 1))
		}
	} else {
		x.Postings[term] = []uint64{x.encode(docId, 1)}
	}
}
