package docache

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

var now = time.Now()

var d1 = Data[int]{Timestamp: now}
var d2 = Data[int]{Timestamp: now.Add(time.Second)}
var d3 = Data[int]{Timestamp: now.Add(2 * time.Second)}

func unsorted() []Data[int] {
	return []Data[int]{d2, d3, d1}
}

func sorted() []Data[int] {
	return []Data[int]{d1, d2, d3}
}

func reversed() []Data[int] {
	return []Data[int]{d3, d2, d1}
}

func before(t *testing.T) (u, s, r []Data[int]) {
	u = unsorted()
	s = sorted()
	r = reversed()

	if reflect.DeepEqual(u, s) {
		t.Fatal("unsorted and sorted []Data[int]'s were deep equal")
	}
	if reflect.DeepEqual(u, r) {
		t.Fatal("unsorted and reversed []Data[int]'s were deep equal")
	}
	if reflect.DeepEqual(s, r) {
		t.Fatal("sorted and reversed []Data[int]'s were deep equal")
	}

	return
}

func TestSort(t *testing.T) {
	u, s, _ := before(t)

	sort.Sort(ByTime[int](u))

	if !reflect.DeepEqual(s, u) {
		t.Fatal("after sort.Sort: unsorted and sorted []Data[int]'s were not deep equal")
	}
}

func TestReverse(t *testing.T) {
	u, _, r := before(t)

	sort.Sort(sort.Reverse(ByTime[int](u)))

	if !reflect.DeepEqual(r, u) {
		t.Fatal("after sort.Reverse: unsorted and reversed []Data[int]'s were not deep equal")
	}
}
