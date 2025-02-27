package proto

import "sort"

func (x *RepeatedValue) Len() int           { return len(x.Ints) }
func (x *RepeatedValue) Less(i, j int) bool { return x.Ints[i] < x.Ints[j] }
func (x *RepeatedValue) Swap(i, j int)      { x.Ints[i], x.Ints[j] = x.Ints[j], x.Ints[i] }

// SortInts is a convenience method: x.Sort() calls Sort(x).
func (x *RepeatedValue) SortInts() { sort.Sort(x) }
