package keystone

import (
	"fmt"
	"github.com/keystonedb/sdk-go/proto"
	"strings"
)

type entityConverter struct {
	protoResponse *proto.EntityResponse
}

func (e *entityConverter) SetProperties(props map[Property]*proto.Value) {
	if e.protoResponse.Properties == nil {
		e.protoResponse.Properties = make([]*proto.EntityProperty, 0)
	}
	for k, v := range props {
		e.protoResponse.Properties = append(e.protoResponse.Properties, &proto.EntityProperty{
			Property: k.Name(),
			Value:    v,
		})
	}
}

func (e *entityConverter) KeystoneProperties() map[Property]*proto.Value {
	resp := make(map[Property]*proto.Value)

	if e.protoResponse.Entity != nil {
		er := e.protoResponse.Entity
		split := strings.Split(er.GetEntityId(), "-")
		resp[knownProperty("_entity_id")] = valueFromString(split[0])
		if len(split) == 2 {
			resp[knownProperty("_child_id")] = valueFromString(split[1])
		}

		resp[knownProperty("_schema_id")] = valueFromString(er.GetSchemaId())
		resp[knownProperty("_created")] = valueFromAny(er.GetCreated())
		resp[knownProperty("_state_change")] = valueFromAny(er.GetStateChange())
		resp[knownProperty("_state")] = valueFromAny(er.GetState())
		resp[knownProperty("_last_update")] = valueFromAny(er.GetLastUpdate())
	}

	var countReplace = map[string]int64{}

	if e.protoResponse.GetRelationshipCounts() != nil {
		for _, v := range e.protoResponse.GetRelationshipCounts() {
			t := v.GetType()
			cnt := int64(v.GetCount())
			if t.GetKey() == "" {
				countReplace["_count_relation"] = cnt
			} else {
				countReplace[fmt.Sprintf("_count_relation:%s:%s:%s", t.GetSource().GetVendorId(), t.GetSource().GetAppId(), t.GetKey())] = cnt
				countReplace[fmt.Sprintf("_count_relation:%s:%s", t.GetSource().GetAppId(), t.GetKey())] = cnt
				countReplace[fmt.Sprintf("_count_relation:%s", t.GetKey())] = cnt
			}
		}
	}

	if e.protoResponse.GetDescendantCounts() != nil {
		for _, v := range e.protoResponse.GetDescendantCounts() {
			t := v.GetType()
			cnt := int64(v.GetCount())
			if t.GetKey() == "" {
				countReplace["_count_descendant"] = cnt
			} else {
				countReplace[fmt.Sprintf("_count_descendant:%s:%s:%s", t.GetSource().GetVendorId(), t.GetSource().GetAppId(), t.GetKey())] = cnt
				countReplace[fmt.Sprintf("_count_descendant:%s:%s", t.GetSource().GetAppId(), t.GetKey())] = cnt
				countReplace[fmt.Sprintf("_count_descendant:%s", t.GetKey())] = cnt
			}
		}
	}

	for variant, cnt := range countReplace {
		resp[knownProperty(variant)] = valueFromInt(cnt)
	}

	return resp
}

func (e *entityConverter) Properties() map[Property]*proto.Value {
	resp := e.KeystoneProperties()
	if e.protoResponse.Properties == nil {
		return resp
	}
	for _, v := range e.protoResponse.Properties {
		nameP := strings.SplitN(v.GetProperty(), ".", 2)
		var prop Property
		if len(nameP) == 2 {
			prop = NewPrefixProperty(nameP[0], nameP[1])
		} else {
			prop = NewProperty(nameP[0])
		}
		resp[prop] = v.GetValue()
	}
	return resp
}
