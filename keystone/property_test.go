package keystone

import "testing"

func Test_PropertyNils(t *testing.T) {
	var prop *Property
	if prop.Name() != "" {
		t.Errorf("Property.Name() = %s; want empty string", prop.Name())
	}

	// Should not panic
	prop.SetPrefix("")
}

func Test_Property_Names(t *testing.T) {
	tests := []struct {
		name     string
		property []string
		expect   []string
	}{
		{"empty", []string{}, []string{}},
		{"single", []string{"TestOne"}, []string{"test_one"}},
		{"multiple", []string{"TestTwo", "TestThree"}, []string{"test_two", "test_three"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			names := PropertyNames(test.property...)
			if len(names) != len(test.expect) {
				t.Errorf("PropertyNames() = %v; want %v", names, test.expect)
			}
			for i, name := range names {
				if name != test.expect[i] {
					t.Errorf("PropertyNames()[%d] = %s; want %s", i, name, test.expect[i])
				}
			}
		})
	}
}

func Test_Property_Name(t *testing.T) {
	tests := []struct {
		name     string
		property string
		expect   string
	}{
		{"PIIToken", "PIIToken", "pii_token"},
		{"snake_case", "snake_case", "snake_case"},
		{"camelCase", "camelCase", "camel_case"},
		{"PascalCase", "PascalCase", "pascal_case"},
		{"kebab-case", "kebab-case", "kebab_case"},
		{"UPPERCASE", "UPPERCASE", "uppercase"},
		{"lowercase", "lowercase", "lowercase"},
		{"KeystoneIDs", "KeystoneIDs", "keystone_ids"},
		{"http2test", "http2test", "http_2_test"},
		{"http2Test", "http2Test", "http_2_test"},
		{"Http2test", "Http2test", "http_2_test"},
		{"Http2Test", "Http2Test", "http_2_test"},
		{"HTTP2Test", "HTTP2Test", "http_2_test"},
		{"HTTP2TEST", "HTTP2TEST", "http_2_test"},
		{"With3dsData", "With3dsData", "with_3_ds_data"},
		{"test123test", "test123test", "test_123_test"},
		{"Test123Test", "Test123Test", "test_123_test"},
		{"KeystoneIDsToUse", "KeystoneIDsToUse", "keystone_ids_to_use"},
		{"Line1", "Line1", "line1"},
		{"Last4", "Last4", "last4"},
		{"CardLast4", "CardLast4", "card_last4"},
		{"ThreeDSecureID", "ThreeDSecureID", "three_d_secure_id"},
		{"TDSecureID", "TDSecureID", "td_secure_id"},
		{"ThreeHTTP", "ThreeHTTP", "three_http"},
		{"ThreeHTTP2", "ThreeHTTP2", "three_http2"},
		{"ThreeHTTP2Score", "ThreeHTTP2Score", "three_http_2_score"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prop := NewProperty(test.property)
			if prop.Name() != test.expect {
				t.Errorf("Property.Name() = %s; want %s", prop.Name(), test.expect)
			}
		})
	}
}

func Test_Property_SetPrefix(t *testing.T) {
	tests := []struct {
		name     string
		property string
		prefix   string
		expect   string
	}{
		{"snake_case", "snake_case", "prefix", "prefix.snake_case"},
		{"camelCase", "camelCase", "prefix", "prefix.camel_case"},
		{"PascalCase", "PascalCase", "prefix", "prefix.pascal_case"},
		{"kebab-case", "kebab-case", "prefix", "prefix.kebab_case"},
		{"UPPERCASE", "UPPERCASE", "prefix", "prefix.uppercase"},
		{"lowercase", "lowercase", "prefix", "prefix.lowercase"},
		{"secondary prefix", "lowercase", "prefix.secondary", "prefix.secondary.lowercase"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prop := NewProperty(test.property)
			prop.SetPrefix(test.prefix)
			if prop.Name() != test.expect {
				t.Errorf("Property.Name() = %s; want %s", prop.Name(), test.expect)
			}
		})
	}
}

func Test_Type(t *testing.T) {
	tests := []struct {
		name   string
		with   interface{}
		expect string
	}{
		{"MarshaledEntity", MarshaledEntity{}, "marshaled-entity"},
		{"MarshaledEntity", &MarshaledEntity{}, "marshaled-entity"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if Type(test.with) != test.expect {
				t.Errorf("Type() = %s; want %s", Type(test.with), test.expect)
			}
		})
	}
}
