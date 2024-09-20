package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"time"
)

// GenericResult is a map that can be used to retrieve a generic result
type GenericResult map[string]interface{}

func makeEntityPropertyMap(resp *proto.EntityResponse) map[string]*proto.EntityProperty {
	entityPropertyMap := map[string]*proto.EntityProperty{}
	for _, p := range resp.GetProperties() {
		entityPropertyMap[p.Property] = p
	}
	return entityPropertyMap
}

func UnmarshalGeneric(resp *proto.EntityResponse, dst GenericResult) error {
	entityPropertyMap := makeEntityPropertyMap(resp)
	for _, p := range entityPropertyMap {

		// Handle Amounts
		if p.Value.GetText() != "" && p.Value.GetInt() > 0 {
			dst[p.Property] = NewAmount(p.Value.GetText(), p.Value.GetInt())
		}

		// Handle Secret text
		if p.Value.GetSecureText() != "" && p.Value.GetText() != "" {
			dst[p.Property] = NewSecretString(p.Value.GetSecureText(), p.Value.GetText())
		}

		if p.Value.GetText() != "" {
			dst[p.Property] = p.Value.GetText()
		}
		if p.Value.GetSecureText() != "" {
			dst[p.Property] = p.Value.GetSecureText()
		}

		if p.Value.GetInt() != 0 {
			dst[p.Property] = p.Value.GetInt()
		}
		if p.Value.GetBool() {
			dst[p.Property] = p.Value.GetBool()
		}
		if p.Value.GetFloat() != 0 {
			dst[p.Property] = p.Value.GetFloat()
		}
		if len(p.Value.GetRaw()) != 0 {
			dst[p.Property] = p.Value.GetRaw()
		}
		if len(p.GetValue().GetArray().GetStrings()) > 0 {
			dst[p.Property] = p.GetValue().GetArray().GetStrings()
		}
		if len(p.GetValue().GetArray().GetInts()) > 0 {
			dst[p.Property] = p.GetValue().GetArray().GetInts()
		}
		if len(p.GetValue().GetArray().GetKeyValue()) > 0 {
			dst[p.Property] = p.GetValue().GetArray().GetKeyValue()
		}
		if p.Value.GetTime() != nil {
			dst[p.Property] = time.Unix(p.Value.GetTime().Seconds, int64(p.Value.GetTime().Nanos))
		}
	}
	return nil
}
