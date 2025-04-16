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

func NewID(parent, child string) ID {
	if child == "" {
		return ID(parent)
	}
	return ID(parent + "-" + child)
}

func assertHashID(input string) {
	if strings.Contains(input, "#") {
		panic("keystone HashID input cannot contain #")
	}
}

func HashID(input string) ID {
	assertHashID(input)
	return ID("#" + input + "#")
}

func HashCID(input, child string) ID {
	if child == "" {
		return HashID(input)
	}
	assertHashID(input)
	return ID("#" + input + "#-" + child)
}
