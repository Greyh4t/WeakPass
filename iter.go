package main

import (
	"math"
	"sync"
)

type Iter struct {
	x       int
	Total   int
	keyList []string
	lists   [][]string
	w       sync.Mutex
}

func (iter *Iter) Init(listsMap map[string][]string, keyList []string) {
	iter.Total = 1
	lists := [][]string{}
	for _, key := range keyList {
		iter.Total *= len(listsMap[key])
		lists = append(lists, listsMap[key])
	}
	iter.keyList = keyList
	iter.lists = lists
}

func (iter *Iter) Percent() float64 {
	return math.Trunc((float64(iter.x*100)/float64(iter.Total)+0.5/math.Pow10(2))*math.Pow10(2)) / math.Pow10(2)
}

func (iter *Iter) Next() map[string]string {
	iter.w.Lock()
	defer iter.w.Unlock()
	step := iter.Total
	if iter.x >= iter.Total {
		return nil
	}
	item := map[string]string{}
	for i, l := range iter.lists {
		step /= len(l)
		item[iter.keyList[i]] = l[iter.x/step%len(l)]
	}
	iter.x += 1
	return item
}

func (iter *Iter) Reset() {
	iter.x = 0
}
