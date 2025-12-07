package keystone

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/keystonedb/sdk-go/proto"
	"github.com/packaged/helpers-go"
)

// Translations is a map of language codes to translated text
type Translations struct {
	values          map[string]*Translation
	toAdd           map[string]*Translation
	toRemove        map[string]bool
	replaceExisting bool
}

func (t *Translations) prepare() {
	if t.toAdd == nil {
		t.toAdd = make(map[string]*Translation)
	}
	if t.toRemove == nil {
		t.toRemove = make(map[string]bool)
	}
	if t.values == nil {
		t.values = make(map[string]*Translation)
	}
}

// Replace replaces all translations with the provided map
func (t *Translations) Replace(translations map[string]*Translation) {
	t.prepare()
	t.values = make(map[string]*Translation)
	t.toAdd = make(map[string]*Translation)
	t.toRemove = make(map[string]bool)
	t.replaceExisting = true
	for lang, text := range translations {
		t.values[lang] = text
	}
}

// Add adds or updates a translation for the given language
func (t *Translations) Add(language string, singular string) {
	t.AddT(language, &Translation{Singular: singular})
}

func (t *Translations) AddT(language string, text *Translation) {
	t.prepare()
	t.toAdd[language] = text
	delete(t.toRemove, language)
}

// Remove removes a translation for the given language
func (t *Translations) Remove(language string) {
	t.prepare()
	t.toRemove[language] = true
	delete(t.toAdd, language)
}

// Get returns the translation for the given language and whether it exists
func (t *Translations) Get(language string) (*Translation, bool) {
	t.prepare()
	if text, ok := t.toAdd[language]; ok {
		return text, true
	}
	if _, ok := t.toRemove[language]; ok {
		return nil, false
	}
	text, ok := t.values[language]
	return text, ok
}

func (t *Translations) FallbackLang(language string, fallbackLang string) *Translation {
	t.prepare()
	if res, ok := t.Get(language); ok {
		return res
	}
	if res, ok := t.Get(fallbackLang); ok {
		return res
	}
	return nil
}

func (t *Translations) Fallback(language, text string) *Translation {
	t.prepare()
	if res, ok := t.Get(language); ok {
		return res
	}
	return &Translation{Singular: text, Plural: text}
}

// All returns all current translations
func (t *Translations) All() map[string]*Translation {
	t.prepare()
	all := make(map[string]*Translation)
	for lang, text := range t.values {
		if _, ok := t.toRemove[lang]; !ok {
			all[lang] = text
		}
	}
	for lang, text := range t.toAdd {
		all[lang] = text
	}
	return all
}

func (t *Translations) applyValues(with map[string]*Translation) {
	t.prepare()
	t.values = with
}

func (t *Translations) merge() {
	useVals := t.All()
	t.values = nil
	t.toAdd = nil
	t.toRemove = nil
	t.replaceExisting = false
	t.prepare()
	t.values = useVals
}

func (t *Translations) MarshalValue() (*proto.Value, error) {
	t.prepare()
	val := &proto.Value{}
	val.Array = proto.NewRepeatedValue()

	// Convert map[string]string to map[string][]byte for KeyValue
	for lang, text := range t.values {
		jsn, _ := json.Marshal(text)
		val.Array.KeyValue[lang] = jsn
	}

	if len(t.toAdd) > 0 {
		val.ArrayAppend = proto.NewRepeatedValue()
		for lang, text := range t.toAdd {
			jsn, _ := json.Marshal(text)
			val.ArrayAppend.KeyValue[lang] = jsn
		}
	}

	if len(t.toRemove) > 0 {
		val.ArrayReduce = proto.NewRepeatedValue()
		for lang := range t.toRemove {
			val.ArrayReduce.KeyValue[lang] = nil
		}
	}

	return val, nil
}

func (t *Translations) UnmarshalValue(value *proto.Value) error {
	if value != nil {
		if value.Array != nil && value.Array.GetKeyValue() != nil {
			newVal := make(map[string]*Translation)
			for lang, textBytes := range value.Array.GetKeyValue() {
				tran := &Translation{}
				if err := tran.fromRaw(textBytes); err == nil {
					newVal[lang] = tran
				}
			}
			t.applyValues(newVal)
		}
		if value.ArrayAppend != nil && value.ArrayAppend.GetKeyValue() != nil {
			for lang, textBytes := range value.ArrayAppend.GetKeyValue() {
				tran := &Translation{}
				if err := tran.fromRaw(textBytes); err == nil {
					t.AddT(lang, tran)
				}
			}
		}
		if value.ArrayReduce != nil && value.ArrayReduce.GetKeyValue() != nil {
			for lang := range value.ArrayReduce.GetKeyValue() {
				t.Remove(lang)
			}
		}
	}
	return nil
}

func (t *Translations) IsZero() bool {
	return t == nil ||
		(len(t.values) == 0 &&
			len(t.toAdd) == 0 &&
			len(t.toRemove) == 0)
}

func (t *Translations) PropertyDefinition() proto.PropertyDefinition {
	return proto.PropertyDefinition{DataType: proto.Property_KeyValue}
}

func (t *Translations) ObserveMutation(resp *proto.MutateResponse) {
	if resp.GetSuccess() {
		t.merge()
	}
}

func (t *Translations) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.All())
}
func (t *Translations) UnmarshalJSON(data []byte) error {
	newMap := make(map[string]*Translation)
	if err := json.Unmarshal(data, &newMap); err != nil {
		return err
	}
	t.Replace(newMap)
	return nil
}

type Translation struct {
	Singular string `json:"s,omitempty"`
	Plural   string `json:"p,omitempty"`
}

func NewTranslation(input ...string) *Translation {
	t := &Translation{}
	switch len(input) {
	case 0:
		return t
	case 1:
		t.Singular = input[0]
	default:
		t.Singular = input[0]
		t.Plural = input[1]
	}
	return t
}

func (t *Translation) fromRaw(data []byte) error {
	if len(data) == 0 {
		return nil
	}
	if string(data[0]) == "{" {
		return json.Unmarshal(data, t)
	} else {
		t.Singular = string(data)
	}
	return nil
}

func (t *Translation) String() string {
	return t.Singular
}

func (t *Translation) GetPlural(quantity int64) string {
	if quantity != 1 && t.Plural == "" {
		t.Plural = strings.ReplaceAll(t.Plural, "(s)", helpers.If(quantity == 1, "", "s"))
	}
	return t.Plural
}

func (t *Translation) Replacements(original string, args map[string]interface{}) string {
	if args == nil {
		return original
	}
	for k, v := range args {
		original = strings.ReplaceAll(original, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return original
}
