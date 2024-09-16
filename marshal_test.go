package keystone

import (
	"errors"
	"github.com/keystonedb/sdk-go/sdk-go/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

type baseE struct {
	ID string
}

type subStruct struct {
	SubName string
}

type marshal struct {
	xName string
}

func (m marshal) MarshalKeystone() (map[Property]*proto.Value, error) {
	result := NewMarshaledEntity()

	if err := result.Append("name", m.xName); err != nil {
		return nil, err
	}

	return result.Properties, nil
}

type selfMarshal struct {
	privateName string
	privateAge  int
	hiddenBool  bool
}

func (s *selfMarshal) MarshalKeystone() (map[Property]*proto.Value, error) {
	result := NewMarshaledEntity()

	if err := result.Append("name", s.privateName); err != nil {
		return nil, err
	}
	if err := result.Append("age", s.privateAge); err != nil {
		return nil, err
	}
	if err := result.Append("hidden", s.hiddenBool); err != nil {
		return nil, err
	}

	return result.Properties, nil
}

type testEntity struct {
	baseE
	Name       string
	Age        int
	DOB        time.Time
	HasGlasses bool
	FraudScore float64
	AmountPaid Amount
	SubStruct  subStruct
}

func Test_MarshalNil(t *testing.T) {
	res, err := Marshal(nil)
	if res != nil {
		t.Errorf("Marshal returned non-nil result")
	}
	if !errors.Is(err, CannotMarshalNil) {
		t.Errorf("Marshal failed: %v", err)
	}
}

func Test_MarshalPrimitives(t *testing.T) {
	tests := []struct {
		name string
		val  interface{}
	}{
		{"string", "Hello, World!"},
		{"bool", true},
		{"int", 42},
		{"int8", int8(42)},
		{"int16", int16(4233)},
		{"int32", int32(422332)},
		{"int64", int64(4223973223)},
		{"uint", uint(42)},
		{"uint8", uint8(42)},
		{"uint16", uint16(4233)},
		{"uint32", uint32(422332)},
		{"uint64", uint64(4223973223)},
		{"float32", float32(42.42)},
		{"float64", float64(42.42)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := Marshal(test.val)
			if !errors.Is(err, CannotMarshalPrimitives) {
				t.Errorf("Marshal failed: %v", err)
			}
			if res != nil {
				t.Errorf("Marshal returned non-nil result")
			}
		})
	}
}

func errorIf(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func Test_SelfMarshal(t *testing.T) {
	s := selfMarshal{
		privateName: "John",
		privateAge:  30,
		hiddenBool:  true,
	}
	props, err := Marshal(s)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	errorIf(t, proto.MatchValue(props[NewProperty("name")], "name", &proto.Value{Text: "John"}))
	errorIf(t, proto.MatchValue(props[NewProperty("age")], "age", &proto.Value{Int: 30}))
	errorIf(t, proto.MatchValue(props[NewProperty("hidden")], "hidden", &proto.Value{Bool: true}))

	propsPoint, err := Marshal(&s)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	errorIf(t, proto.MatchValue(propsPoint[NewProperty("name")], "name", &proto.Value{Text: "John"}))
	errorIf(t, proto.MatchValue(propsPoint[NewProperty("age")], "age", &proto.Value{Int: 30}))
	errorIf(t, proto.MatchValue(propsPoint[NewProperty("hidden")], "hidden", &proto.Value{Bool: true}))
}

func Test_SelfMarshalNonPointer(t *testing.T) {
	s := marshal{
		xName: "Harry",
	}
	props, err := Marshal(s)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	errorIf(t, proto.MatchValue(props[NewProperty("name")], "name", &proto.Value{Text: "Harry"}))

	propsPoint, err := Marshal(&s)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	errorIf(t, proto.MatchValue(propsPoint[NewProperty("name")], "name", &proto.Value{Text: "Harry"}))
}

func TestMarshal(t *testing.T) {
	amnt := NewAmount("USD", 142)
	e := testEntity{
		baseE:      baseE{ID: "xx"},
		Name:       "John",
		Age:        30,
		DOB:        time.Date(2009, 2, 13, 23, 31, 30, 12345, time.UTC),
		HasGlasses: true,
		FraudScore: 0.5,
		AmountPaid: amnt,
		SubStruct:  subStruct{SubName: "AAA"},
	}
	props, err := Marshal(e)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	errorIf(t, proto.MatchValue(props[NewProperty("ID")], "ID", &proto.Value{Text: "xx"}))
	errorIf(t, proto.MatchValue(props[NewProperty("Name")], "Name", &proto.Value{Text: "John"}))
	errorIf(t, proto.MatchValue(props[NewProperty("Age")], "Age", &proto.Value{Int: 30}))
	errorIf(t, proto.MatchValue(props[NewProperty("DOB")], "DOB", &proto.Value{Time: timestamppb.New(time.Date(2009, 2, 13, 23, 31, 30, 12345, time.UTC))}))
	errorIf(t, proto.MatchValue(props[NewProperty("HasGlasses")], "HasGlasses", &proto.Value{Bool: true}))
	errorIf(t, proto.MatchValue(props[NewProperty("FraudScore")], "FraudScore", &proto.Value{Float: 0.5}))
	errorIf(t, proto.MatchValue(props[NewProperty("AmountPaid")], "AmountPaid", &proto.Value{Text: "USD", Int: 142}))

	errorIf(t, proto.MatchValue(props[NewPrefixProperty("sub_struct", "SubName")], "SubName", &proto.Value{Text: "AAA"}))
}

func TestMarshalPointer(t *testing.T) {
	amnt := NewAmount("USD", 142)
	e := &testEntity{
		baseE:      baseE{ID: "xx"},
		Name:       "John",
		Age:        30,
		DOB:        time.Date(2009, 2, 13, 23, 31, 30, 12345, time.UTC),
		HasGlasses: true,
		FraudScore: 0.5,
		AmountPaid: amnt,
		SubStruct:  subStruct{SubName: "AAA"},
	}
	props, err := Marshal(e)
	if err != nil {
		t.Errorf("Marshal failed: %v", err)
	}

	errorIf(t, proto.MatchValue(props[NewProperty("ID")], "ID", &proto.Value{Text: "xx"}))
	errorIf(t, proto.MatchValue(props[NewProperty("Name")], "Name", &proto.Value{Text: "John"}))
	errorIf(t, proto.MatchValue(props[NewProperty("Age")], "Age", &proto.Value{Int: 30}))
	errorIf(t, proto.MatchValue(props[NewProperty("DOB")], "DOB", &proto.Value{Time: timestamppb.New(time.Date(2009, 2, 13, 23, 31, 30, 12345, time.UTC))}))
	errorIf(t, proto.MatchValue(props[NewProperty("HasGlasses")], "HasGlasses", &proto.Value{Bool: true}))
	errorIf(t, proto.MatchValue(props[NewProperty("FraudScore")], "FraudScore", &proto.Value{Float: 0.5}))
	errorIf(t, proto.MatchValue(props[NewProperty("AmountPaid")], "AmountPaid", &proto.Value{Text: "USD", Int: 142}))
}

func TestMarshal_NestedFailure(t *testing.T) {
	xErr := errors.New("expect error")
	x := struct {
		Name      string
		SubStruct struct {
			Issue testValueMarshaler
		}
	}{
		Name: "James",
		SubStruct: struct {
			Issue testValueMarshaler
		}{
			testValueMarshaler{error: xErr},
		},
	}

	resp, err := Marshal(x)
	if !errors.Is(err, xErr) {
		t.Errorf("Expected error, got: %v", err)
	}
	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}
}

func Test_MarshaledEntity_Append(t *testing.T) {
	m := NewMarshaledEntity()
	err := m.Append("name", "John")
	if err != nil {
		t.Errorf("Append failed: %v", err)
	}

	xChan := make(chan string)
	err = m.Append("chan", xChan)
	if !errors.Is(err, CannotMarshalValueError) {
		t.Errorf("Expected CannotMarshalValueError, got: %v", err)
	}

	expectErr := errors.New("expect err")
	x := &testValueMarshaler{error: expectErr}
	xErr := m.Append("test", x)
	if !errors.Is(xErr, expectErr) {
		t.Errorf("Expected error, got: %v", xErr)
	}
}

func Test_DynamicStructErr(t *testing.T) {
	expectErr := errors.New("expect err")
	x := struct {
		Name    string
		TestVal testValueMarshaler
	}{
		Name:    "Testing",
		TestVal: testValueMarshaler{error: expectErr},
	}

	resp, err := Marshal(x)
	if !errors.Is(err, expectErr) {
		t.Errorf("Expected error, got: %v", err)
	}
	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}
}
