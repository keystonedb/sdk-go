package keystone

import "strings"

// ID is a unique identifier for a remote object
type ID string

func (id ID) String() string {
	return string(id)
}

func (id ID) ParentID() string {
	split := strings.SplitN(string(id), "-", 2)
	return split[0]
}

func (id ID) ChildID() string {
	split := strings.SplitN(string(id), "-", 2)
	if len(split) > 1 {
		return split[1]
	}
	return ""
}

func (id ID) Matches(input string) bool {
	return string(id) == input
}
