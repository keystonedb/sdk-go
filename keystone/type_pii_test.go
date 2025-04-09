package keystone

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBasicMask(t *testing.T) {
	tests := []struct {
		text          string
		expectedMatch *regexp.Regexp
	}{
		{"Tom Kay", regexp.MustCompile("T\\*+m K\\*+y")},
		{"Testing Út 106 5", regexp.MustCompile("T\\*+g Ú\\*+ 1\\*+6 \\*+")},
	}

	for _, test := range tests {
		t.Run(test.text, func(t *testing.T) {
			masked := BasicMask(test.text)
			assert.NotEqual(t, test.text, masked)
			assert.Regexp(t, test.expectedMatch, masked)
		})
	}
}

func TestMaskEmail(t *testing.T) {
	tests := []struct {
		email         string
		expectedMatch *regexp.Regexp
	}{
		{"tom@tomkay.me", regexp.MustCompile("t\\*+m@t\\*+y.me")},
		{"tom.kay@chargehive.com", regexp.MustCompile("t\\*+y@chargehive.com")},
	}

	for _, test := range tests {
		t.Run(test.email, func(t *testing.T) {
			secure := NewEmail(test.email)
			assert.Equal(t, test.email, secure.Original)
			assert.Regexp(t, test.expectedMatch, secure.Masked)
		})
	}
}

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		phone         string
		expectedMatch *regexp.Regexp
	}{
		{"+440123456789", regexp.MustCompile("\\+44012\\*+789")},
		{"0010123456789", regexp.MustCompile("001012\\*+789")},
	}

	for _, test := range tests {
		t.Run(test.phone, func(t *testing.T) {
			secure := NewPhone(test.phone)
			assert.Equal(t, test.phone, secure.Original)
			assert.Regexp(t, test.expectedMatch, secure.Masked)
		})
	}
}

func TestMaskIPAddress(t *testing.T) {
	tests := []struct {
		ip               string
		expectedOriginal string
		expectedMasked   string
	}{
		{"", "", ""},
		{"invalid", "", ""},
		{"10.10.10.10", "10.10.10.10", "10.10.10.0"},
	}

	for _, test := range tests {
		t.Run(test.ip, func(t *testing.T) {
			secure := NewSecureIPV4(test.ip)
			assert.Equal(t, test.expectedOriginal, secure.Original)
			assert.Equal(t, test.expectedMasked, secure.Masked)
		})
	}
}
