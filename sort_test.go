package docache

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

var now time.Time = time.Now()

func unsorted() []Data[int] {
	return []Data[int]{
		Data[int]{
			Timestamp: now.Add(time.Second),
		},
		Data[int]{
			Timestamp: now.Add(2 * time.Second),
		},
		Data[int]{
			Timestamp: now,
		},
	}
}

func sorted() []Data[int] {
	return []Data[int]{
		Data[int]{
			Timestamp: now,
		},
		Data[int]{
			Timestamp: now.Add(time.Second),
		},
		Data[int]{
			Timestamp: now.Add(2 * time.Second),
		},
	}
}

func reversed() []Data[int] {
	return []Data[int]{
		Data[int]{
			Timestamp: now.Add(2 * time.Second),
		},
		Data[int]{
			Timestamp: now.Add(time.Second),
		},
		Data[int]{
			Timestamp: now,
		},
	}
}

func before(t *testing.T) {
	if reflect.DeepEqual(unsorted(), sorted()) {
		t.Fatal("unsorted and sorted []Data[int]'s were deep equal")
	}
	if reflect.DeepEqual(unsorted(), reversed()) {
		t.Fatal("unsorted and reversed []Data[int]'s were deep equal")
	}
	if reflect.DeepEqual(sorted(), reversed()) {
		t.Fatal("sorted and reversed []Data[int]'s were deep equal")
	}
}

func TestSort(t *testing.T) {
	before(t)

	unsort := unsorted()

	sort.Sort(ByTime[int](unsort))

	if !reflect.DeepEqual(sorted(), unsort) {
		t.Fatal("after sort.Sort: unsorted and sorted []Data[int]'s were not deep equal")
	}
}

func TestReverse(t *testing.T) {
	before(t)

	unsort := unsorted()

	sort.Sort(sort.Reverse(ByTime[int](unsort)))

	if !reflect.DeepEqual(reversed(), unsort) {
		t.Fatal("after sort.Reverse: unsorted and reversed []Data[int]'s were not deep equal")
	}
}
