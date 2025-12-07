package keystone

import (
	"reflect"
	"regexp"
	"strings"

	"github.com/keystonedb/sdk-go/keystone/reflector"
	"github.com/keystonedb/sdk-go/proto"
)

func knownProperty(name string) Property {
	return Property{name: name}
}

func knownPrefixProperty(prefix, name string) Property {
	return Property{prefix: prefix, name: name}
}

func NewProperty(name string) Property {
	return Property{name: PropertyName(name)}
}

func NewPrefixProperty(prefix, name string) Property {
	return Property{prefix: prefix, name: PropertyName(name)}
}

type Property struct {
	prefix string
	name   string
}

func (p *Property) HydrateOnly() bool {
	return p != nil && strings.HasPrefix(p.name, "_")
}

func (p *Property) SetPrefix(prefix string) {
	if p != nil {
		p.prefix = prefix
	}
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

var matchWords = []*regexp.Regexp{
	regexp.MustCompile("([0-9]+)([A-Za-z])"),
	regexp.MustCompile("([A-Z][a-z0-9]{2,})"),
	regexp.MustCompile("([A-Z]+[a-z])([^a-z]|$)"),
}
var matchNonAlphaNum = regexp.MustCompile("([^a-z0-9A-Z])")

func PrefixedPropertyNames(prefix string, str ...string) []string {
	names := make([]string, len(str))
	for i, s := range str {
		names[i] = prefix + PropertyName(s)
	}
	return names
}

func PropertyNames(str ...string) []string {
	names := make([]string, len(str))
	for i, s := range str {
		names[i] = PropertyName(s)
	}
	return names
}

func PropertyName(str string) string {
	for _, match := range matchWords {
		str = match.ReplaceAllString(str, "_${1}_${2}")
	}
	str = matchNonAlphaNum.ReplaceAllString(str, "_")
	// trim leading and trailing underscores
	str = strings.Trim(str, "_")
	// remove double underscores
	str = strings.ReplaceAll(str, "__", "_")
	// lowercase
	return strings.ToLower(str)
}

func Type(input interface{}) string {
	t := reflector.Deref(reflect.ValueOf(input)).Type()
	return strings.ReplaceAll(PropertyName(t.Name()), "_", "-")
}

func ReflectProperty(f reflect.StructField, prefix string) (Property, proto.PropertyDefinition) {
	opt := getFieldOptions(f)
	return knownPrefixProperty(prefix, opt.name), opt.definition()
}

func mergeDefinitions(a, b proto.PropertyDefinition) proto.PropertyDefinition {
	result := a
	if b.DataType > a.DataType {
		result.DataType = b.DataType
	}
	if b.ExtendedType > a.ExtendedType {
		result.ExtendedType = b.ExtendedType
	}

	if len(b.Options) > 0 {
		if len(a.Options) == 0 {
			result.Options = b.Options
		} else {
			opts := map[proto.Property_Option]bool{}
			for _, opt := range a.Options {
				opts[opt] = true
			}
			for _, opt := range b.Options {
				opts[opt] = true
			}
			result.Options = nil
			for opt := range opts {
				result.Options = append(result.Options, opt)
			}
		}
	}

	return result
}

func getFieldOptions(f reflect.StructField) fieldOptions {
	tag := f.Tag.Get("keystone")
	opt := fieldOptions{}

	tagParts := strings.Split(tag, ",")
	for i, part := range tagParts {
		part = strings.TrimSpace(part)
		if i == 0 {
			if part == "" {
				opt.name = PropertyName(f.Name)
			} else if part == "-" {
				return opt
			} else {
				opt.name = strings.ToLower(part)
			}
			continue
		}
		switch part {
		case "omitempty":
			opt.omitempty = true

		case "unique":
			opt.unique = true
		case "primary":
			opt.primary = true
		case "indexed", "query":
			opt.indexed = true
		case "searchable", "search":
			opt.searchable = true
		case "immutable":
			opt.immutable = true
		case "deprecated":
			opt.deprecated = true
		case "required", "req":
			opt.required = true
		case "lookup":
			opt.reverseLookup = true
		case "verify":
			opt.verifyOnly = true
		case "metric":
			opt.metric = true
		case "metricFilter":
			opt.metricFilter = true
		case "no-snapshot", "skip-snapshot":
			opt.noSnapshot = true

		case "pii", "personal", "gdpr":
			opt.personalData = true
		case "user":
			opt.userInputData = true
		}
	}
	return opt
}

type fieldOptions struct {
	name string

	// marshal
	omitempty bool

	// options
	unique        bool
	primary       bool
	indexed       bool
	searchable    bool
	immutable     bool
	deprecated    bool
	required      bool
	reverseLookup bool
	verifyOnly    bool

	metric       bool
	metricFilter bool
	noSnapshot   bool

	// Data classification
	personalData  bool
	userInputData bool
}

func (fOpt fieldOptions) definition() proto.PropertyDefinition {
	return fOpt.applyTo(proto.PropertyDefinition{})
}

func (fOpt fieldOptions) applyTo(onto proto.PropertyDefinition) proto.PropertyDefinition {
	onto.Options = fOpt.applyOptions(onto.Options)
	onto = fOpt.applyTypes(onto)
	return onto
}

func (fOpt fieldOptions) applyTypes(onto proto.PropertyDefinition) proto.PropertyDefinition {
	if fOpt.personalData {
		onto.ExtendedType = proto.Property_Personal
	} else if fOpt.userInputData {
		onto.ExtendedType = proto.Property_UserInput
	} else if fOpt.verifyOnly {
		onto.DataType = proto.Property_VerifyText
	}
	return onto
}

func (fOpt fieldOptions) applyOptions(onto []proto.Property_Option) []proto.Property_Option {
	onto = appendOption(proto.Property_Unique, onto, fOpt.unique)
	onto = appendOption(proto.Property_Primary, onto, fOpt.primary)
	onto = appendOption(proto.Property_Indexed, onto, fOpt.indexed)
	onto = appendOption(proto.Property_Immutable, onto, fOpt.immutable)
	onto = appendOption(proto.Property_Deprecated, onto, fOpt.deprecated)
	onto = appendOption(proto.Property_Required, onto, fOpt.required)
	onto = appendOption(proto.Property_ReverseLookup, onto, fOpt.reverseLookup)
	onto = appendOption(proto.Property_Searchable, onto, fOpt.searchable)
	onto = appendOption(proto.Property_Metric, onto, fOpt.metric)
	onto = appendOption(proto.Property_MetricFilter, onto, fOpt.metricFilter)
	onto = appendOption(proto.Property_NoSnapshot, onto, fOpt.noSnapshot)
	return onto
}

func appendOption(option proto.Property_Option, onto []proto.Property_Option, when bool) []proto.Property_Option {
	if when {
		onto = append(onto, option)
	}
	return onto
}
