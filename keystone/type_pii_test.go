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
