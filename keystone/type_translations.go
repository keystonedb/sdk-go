package keystone

import "github.com/keystonedb/sdk-go/proto"

// Translations is a map of language codes to translated text
type Translations struct {
	values          map[string]string
	toAdd           map[string]string
	toRemove        map[string]bool
	replaceExisting bool
}

func (t *Translations) prepare() {
	if t.toAdd == nil {
		t.toAdd = make(map[string]string)
	}
	if t.toRemove == nil {
		t.toRemove = make(map[string]bool)
	}
	if t.values == nil {
		t.values = make(map[string]string)
	}
}

// Replace replaces all translations with the provided map
func (t *Translations) Replace(translations map[string]string) {
	t.prepare()
	t.values = make(map[string]string)
	t.toAdd = make(map[string]string)
	t.toRemove = make(map[string]bool)
	t.replaceExisting = true
	for lang, text := range translations {
		t.values[lang] = text
	}
}

// Add adds or updates a translation for the given language
func (t *Translations) Add(language, text string) {
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
func (t *Translations) Get(language string) (string, bool) {
	t.prepare()
	if text, ok := t.toAdd[language]; ok {
		return text, true
	}
	if _, ok := t.toRemove[language]; ok {
		return "", false
	}
	text, ok := t.values[language]
	return text, ok
}

// All returns all current translations
func (t *Translations) All() map[string]string {
	t.prepare()
	all := make(map[string]string)
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

func (t *Translations) applyValues(with map[string]string) {
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
		val.Array.KeyValue[lang] = []byte(text)
	}

	if len(t.toAdd) > 0 {
		val.ArrayAppend = proto.NewRepeatedValue()
		for lang, text := range t.toAdd {
			val.ArrayAppend.KeyValue[lang] = []byte(text)
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
			newVal := make(map[string]string)
			for lang, textBytes := range value.Array.GetKeyValue() {
				newVal[lang] = string(textBytes)
			}
			t.applyValues(newVal)
		}
		if value.ArrayAppend != nil && value.ArrayAppend.GetKeyValue() != nil {
			for lang, textBytes := range value.ArrayAppend.GetKeyValue() {
				t.Add(lang, string(textBytes))
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
