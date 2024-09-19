package keystone

import (
	"github.com/keystonedb/sdk-go/proto"
	"testing"
	"time"
)

type testTimeSeriesEntity struct {
	TimeSeriesEntity
}

func TestTimeSeriesEntity_Definition(t *testing.T) {
	res := Define(&testTimeSeriesEntity{})
	if res.KeystoneType != proto.Schema_TimeSeries {
		t.Error("Define(TimeSeriesEntity) failed to return the correct KeystoneType")
	}
	res2 := Define(&TimeSeriesEntity{})
	if res2.KeystoneType != proto.Schema_TimeSeries {
		t.Error("Define(TimeSeriesEntity) failed to return the correct KeystoneType")
	}
}

func TestTimeSeriesEntity(t *testing.T) {
	testRunStart := time.Now()
	tse := testTimeSeriesEntity{}

	if tse.GetTimeSeriesInputTime().IsZero() || tse.GetTimeSeriesInputTime().Before(testRunStart) {
		t.Error("TimeSeriesEntity.GetTimeSeriesInputTime() failed to return now as a default")
	}

	testTime := time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)
	tse.SetTimeSeriesInputTime(testTime)
	if tse.GetTimeSeriesInputTime() != testTime {
		t.Error("TimeSeriesEntity.GetTimeSeriesInputTime() failed to return the set time")
	}
}
