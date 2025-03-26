package proto

import (
	"bytes"
	"fmt"
)

func MatchValue(prop *Value, name string, expect *Value) error {
	if prop == nil {
		return fmt.Errorf("property is nil - %s", name)
	}
	if prop.GetText() != expect.GetText() {
		return fmt.Errorf("text expect mismatch: %s != %s", prop.GetText(), expect.GetText())
	}

	if prop.GetInt() != expect.GetInt() {
		return fmt.Errorf("int expect mismatch: %d != %d", prop.GetInt(), expect.GetInt())
	}

	if prop.GetBool() != expect.GetBool() {
		return fmt.Errorf("bool expect mismatch: %t != %t", prop.GetBool(), expect.GetBool())
	}

	if prop.GetFloat() != expect.GetFloat() {
		return fmt.Errorf("float expect mismatch: %f != %f", prop.GetFloat(), expect.GetFloat())
	}

	if prop.GetTime().AsTime() != expect.GetTime().AsTime() {
		return fmt.Errorf("time expect mismatch: %v != %v", prop.GetTime().AsTime(), expect.GetTime().AsTime())
	}

	if prop.GetSecureText() != expect.GetSecureText() {
		return fmt.Errorf("secure text expect mismatch: %s != %s", prop.GetSecureText(), expect.GetSecureText())
	}

	if !bytes.Equal(expect.GetRaw(), prop.GetRaw()) {
		return fmt.Errorf("raw expect mismatch: %v != %v", prop.GetRaw(), expect.GetRaw())
	}

	if err := MatchRepeatedValue(prop.GetArray(), expect.GetArray()); err != nil {
		return err
	}
	if err := MatchRepeatedValue(prop.GetArrayReduce(), expect.GetArrayReduce()); err != nil {
		return err
	}
	if err := MatchRepeatedValue(prop.GetArrayAppend(), expect.GetArrayAppend()); err != nil {
		return err
	}

	return nil
}

func MatchRepeatedValue(input, expect *RepeatedValue) error {
	if input.IsZero() && expect.IsZero() {
		return nil
	}

	if input.IsZero() {
		return fmt.Errorf("input is nil, expect is not")
	}

	if expect.IsZero() {
		return fmt.Errorf("expect is nil, input is not")
	}

	if len(input.KeyValue) != len(expect.KeyValue) {
		return fmt.Errorf("array length mismatch: %d != %d", len(input.KeyValue), len(expect.KeyValue))
	}

	if len(input.Ints) != len(expect.Ints) {
		return fmt.Errorf("array length mismatch: %d != %d", len(input.Ints), len(expect.Ints))
	}

	if len(input.Strings) != len(expect.Strings) {
		return fmt.Errorf("array length mismatch: %d != %d", len(input.Strings), len(expect.Strings))
	}

	if len(input.KeyValue) > 0 {
		for i, v := range expect.KeyValue {
			if !bytes.Equal(input.KeyValue[i], v) {
				return fmt.Errorf("array value mismatch: %v != %v", input.KeyValue[i], v)
			}
		}
	}

	// Match ints and strings contain the same values.  Order does not matter

	if len(input.Ints) > 0 {
		expectMap := make(map[int64]int)
		for _, v := range expect.Ints {
			expectMap[v]++
		}
		inputMap := make(map[int64]int)
		for _, v := range input.Ints {
			inputMap[v]++
		}
		for k, v := range expectMap {
			if inputMap[k] != v {
				return fmt.Errorf("array value mismatch: %v != %v", inputMap[k], v)
			}
		}
	}

	if len(input.Strings) > 0 {
		expectMap := make(map[string]int)
		for _, v := range expect.Strings {
			expectMap[v]++
		}
		inputMap := make(map[string]int)
		for _, v := range input.Strings {
			inputMap[v]++
		}
		for k, v := range expectMap {
			if inputMap[k] != v {
				return fmt.Errorf("array value mismatch: %v != %v", inputMap[k], v)
			}
		}
	}

	return nil
}
