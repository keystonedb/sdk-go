package keystone

import "time"

type TimeSeriesEntity struct {
	_timeSeriesInputTime time.Time
}

func (e *TimeSeriesEntity) SetTimeSeriesInputTime(t time.Time) {
	e._timeSeriesInputTime = t
}

func (e *TimeSeriesEntity) GetTimeSeriesInputTime() time.Time {
	if e._timeSeriesInputTime.IsZero() {
		e._timeSeriesInputTime = time.Now()
	}
	return e._timeSeriesInputTime
}

type TSEntity interface {
	GetTimeSeriesInputTime() time.Time
}
