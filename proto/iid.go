package proto

func (i *IIDResponse) IDCount(key string) int64 {
	if i == nil {
		return 0
	}
	return i.Counts[key]
}
