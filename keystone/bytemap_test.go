package keystone

import (
	"testing"
)

func TestByteMap(t *testing.T) {
	type toTest struct {
		Name           string
		Age            int
		UsuallyCorrect bool
	}

	tt := toTest{
		Name:           "Spring",
		Age:            25,
		UsuallyCorrect: true,
	}

	marsh := ToByteMap(tt)

	dec := toTest{}
	FromByteMap(marsh, &dec)

	if dec.Name != tt.Name {
		t.Errorf("Name not equal")
	}
	if dec.Age != tt.Age {
		t.Errorf("Age not equal")
	}
	if dec.UsuallyCorrect != tt.UsuallyCorrect {
		t.Errorf("UsuallyCorrect not equal")
	}
}
