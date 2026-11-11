package keystone

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/keystonedb/sdk-go/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNamedStringTypesKeepExtendedType(t *testing.T) {
	tests := []struct {
		value    any
		extended proto.Property_ExtendedType
	}{
		{Country("GB"), proto.Property_Country},
		{URL("https://example.com"), proto.Property_URL},
		{PII("secret"), proto.Property_Personal},
		{IPAddress("10.0.0.1"), proto.Property_IP},
		{UserInput("hello"), proto.Property_UserInput},
	}
	for _, test := range tests {
		v := reflect.ValueOf(test.value)
		t.Run(v.Type().Name(), func(t *testing.T) {
			ref := GetReflector(v.Type(), v)
			require.NotNil(t, ref)
			def := ref.PropertyDefinition()
			assert.Equal(t, proto.Property_Text, def.DataType)
			assert.Equal(t, test.extended, def.ExtendedType)

			pv, err := ref.ToProto(v)
			require.NoError(t, err)
			assert.Equal(t, v.String(), pv.GetText())

			onto := reflect.New(v.Type()).Elem()
			require.NoError(t, ref.SetValue(pv, onto))
			assert.Equal(t, test.value, onto.Interface())
		})
	}

	ptr := reflect.TypeFor[*Country]()
	assert.Equal(t, proto.Property_Country, GetReflector(ptr, reflect.Value{}).PropertyDefinition().ExtendedType)
}

type secureFixture struct {
	Phone Phone
	Email Email
}

func TestMarshalSecureStringStates(t *testing.T) {
	props, err := Marshal(secureFixture{
		Phone: Phone{},                                // cleared
		Email: Email{SecureString{Masked: "a***b@c"}}, // hydrated without original
	})
	require.NoError(t, err)

	phone, ok := props[NewProperty("phone")]
	require.True(t, ok, "cleared secure value is marshalled")
	assert.Empty(t, phone.GetText())
	assert.Empty(t, phone.GetSecureText())

	_, ok = props[NewProperty("email")]
	assert.False(t, ok, "masked-only secure value is withheld")
}

func TestWatcherSendsClearedSecureValue(t *testing.T) {
	v := &secureFixture{Phone: NewPhone("+440123456789")}
	w, err := NewWatcher(v)
	require.NoError(t, err)

	v.Phone = NewPhone("")
	changes, err := w.Changes(v, false)
	require.NoError(t, err)
	phone, ok := changes[NewProperty("phone")]
	require.True(t, ok)
	assert.Empty(t, phone.GetSecureText())
	assert.NotContains(t, changes, NewProperty("email"))
}

func TestDefaultWatcherIgnoresEmptySecureValue(t *testing.T) {
	changes, err := ChangesFromDefault(&secureFixture{})
	require.NoError(t, err)
	assert.Empty(t, changes)
}

func TestMasksAreStable(t *testing.T) {
	assert.Equal(t, NewEmail("tom.kay@chargehive.com"), NewEmail("tom.kay@chargehive.com"))
	assert.Equal(t, NewPersonName("Tom Kay"), NewPersonName("Tom Kay"))
	assert.Equal(t, NewPhone("+440123456789"), NewPhone("+440123456789"))

	v := &secureFixture{Email: NewEmail("tom@tomkay.me")}
	w, err := NewWatcher(v)
	require.NoError(t, err)
	v.Email = NewEmail("tom@tomkay.me")
	changes, err := w.Changes(v, false)
	require.NoError(t, err)
	assert.Empty(t, changes)
}

func TestSecureStringJSONRoundTrip(t *testing.T) {
	in := secureFixture{Email: NewEmail("tom@tomkay.me")}
	data, err := json.Marshal(in)
	require.NoError(t, err)

	var out secureFixture
	require.NoError(t, json.Unmarshal(data, &out))
	assert.Equal(t, in.Email.Masked, out.Email.Masked)
	assert.Empty(t, out.Email.Original)

	var legacy Email
	require.NoError(t, json.Unmarshal([]byte(`{"masked":"t***m"}`), &legacy))
	assert.Equal(t, "t***m", legacy.Masked)

	var null Email
	require.NoError(t, json.Unmarshal([]byte(`null`), &null))
	assert.Empty(t, null.Masked)
}
