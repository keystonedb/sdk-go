package keystone

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
