// Copyright 2017, Jonsen Yang.  All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"sort"
)

type MapSorter []MapItem
type MapItem struct {
	Key string
	Val int64
}

func NewMapSorter(m map[string]int64) MapSorter {
	ms := make(MapSorter, 0, len(m))

	for k, v := range m {
		ms = append(ms, MapItem{k, v})
	}

	sort.Sort(ms)

	return ms
}

func (ms MapSorter) Len() int {
	return len(ms)
}

func (ms MapSorter) Less(i, j int) bool {
	return ms[i].Val < ms[j].Val
}

func (ms MapSorter) Swap(i, j int) {
	ms[i], ms[j] = ms[j], ms[i]
}
