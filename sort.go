package docache

// an implementation of sort.Interface for a slice of Data[T]

// by default, Cache.Data will return in order from oldest to newest
// but you can use `sort.Sort(sort.Reverse(ByTime[T](data)))`
// to get the data in newest to oldest order if you prefer
type ByTime[T any] []Data[T]

func (a ByTime[T]) Len() int           { return len(a) }
func (a ByTime[T]) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTime[T]) Less(i, j int) bool { return a[i].Timestamp.Before(a[j].Timestamp) }
