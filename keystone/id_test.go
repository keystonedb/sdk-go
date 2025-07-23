package keystone

import (
	"testing"
	"time"
)

func TestIDTime(t *testing.T) {
	id := ID("8Wb9D1aSWv528IIRxCn")
	res := id.Time().UTC().Truncate(time.Second)
	if res.Unix() != 1753260037 {
		t.Errorf("Expected time to be %s got %s # %d", "2025-07-23 08:40:37", res.Format(time.DateTime), res.Unix())
	}
}
