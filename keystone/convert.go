package keystone

import (
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

func (e *entityConverter) Properties() map[Property]*proto.Value {
	resp := make(map[Property]*proto.Value)
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
