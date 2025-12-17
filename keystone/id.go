package keystone

import (
	"strconv"
	"strings"
	"time"

	"github.com/kubex/k4id"
)

var k7ID = k4id.NewGenerator(k4id.TimeGeneratorNano)

// ID is a unique identifier for a remote object
type ID string

// Time returns the time of the parent ID
func (id ID) Time() time.Time {
	return k7ID.ExtractTime(id.ParentID())
}

// ChildTime returns the time of the child ID if available, and not a custom ID
func (id ID) ChildTime() time.Time {
	if id.ChildID() != "" {
		tid, _ := strconv.ParseInt(id.ChildID(), 36, 64)
		return time.Unix(0, tid)
	}
	return time.Time{}
}

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
