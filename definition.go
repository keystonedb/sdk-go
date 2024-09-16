package keystone

import (
	"regexp"
	"strings"
)

func NewProperty(name string) Property {
	return Property{name: snakeCase(name)}
}

func NewPrefixProperty(prefix, name string) Property {
	return Property{prefix: prefix, name: snakeCase(name)}
}

type Property struct {
	prefix string
	name   string
}

func (p *Property) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *Property) Name() string {
	if p == nil {
		return ""
	}
	if p.prefix != "" {
		return p.prefix + "." + p.name
	}
	return p.name
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
var matchNonAlphaNum = regexp.MustCompile("([^a-z0-9A-Z])")

func snakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	snake = matchNonAlphaNum.ReplaceAllString(snake, "_")
	return strings.ToLower(snake)
}
