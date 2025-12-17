package timeseries

import (
	"context"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/proto"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/packaged/logger/v3/logger"
	"go.uber.org/zap"
)

type Requirement struct {
	conn *keystone.Connection
}

func (d *Requirement) Name() string {
	return "Time Series"
}

func (d *Requirement) Register(conn *keystone.Connection) error {
	d.conn = conn
	conn.RegisterTypes(models.ReportedEvent{})
	return nil
}

func (d *Requirement) Verify(actor *keystone.Actor) []requirements.TestResult {
	return []requirements.TestResult{
		d.record(actor),
		d.chart(actor),
	}
}

func (d *Requirement) record(actor *keystone.Actor) requirements.TestResult {
	var updateErr error
	for i := 0; i < 1; i++ {
		evn := &models.ReportedEvent{}
		evn.Library = "libone"
		switch rand.Int32N(3) {
		case 0:
			evn.Library = "libone"
		case 1:
			evn.Library = "libtwo"
		case 2:
			evn.Library = "libthree"
		case 3:
			evn.Library = "libfour"
		}

		evn.BinData = "To Gather"

		evn.Converted = rand.Int32N(3) > 1

		statusCode := http.StatusNotFound
		switch rand.Int32N(5) {
		case 0:
			statusCode = http.StatusOK
		case 1:
			statusCode = http.StatusInternalServerError
		case 2:
			statusCode = http.StatusBadRequest
		case 3:
			statusCode = http.StatusTemporaryRedirect
		case 4:
			statusCode = http.StatusUnauthorized
		}

		evn.ResultCode = strconv.Itoa(statusCode)

		evn.LibraryID = "lib-id-123"
		switch rand.Int32N(6) {
		case 0:
			evn.LibraryID = "lib-id-345"
		case 1:
			evn.LibraryID = "lib-id-678"
		case 2:
			evn.LibraryID = "lib-id-901"
		case 3:
			evn.LibraryID = "lib-id-234"
		case 4:
			evn.LibraryID = "lib-id-567"
		case 5:
			evn.LibraryID = "lib-id-890"
		}
		evn.ResponseMessage = http.StatusText(statusCode)
		evn.SetTimeSeriesInputTime(time.Now())
		evn.AddLabel("lbl1", "val1")
		evn.AddLabel("lbl2", "val2")
		updateErr = actor.ReportTimeSeries(context.Background(), evn)
	}

	return requirements.TestResult{
		Name:  "Store TimeSeries",
		Error: updateErr,
	}
}

func (d *Requirement) chart(actor *keystone.Actor) requirements.TestResult {

	evn := &models.ReportedEvent{}

	res, chartErr := d.conn.DirectClient().ChartTimeSeries(context.Background(), &proto.ChartTimeSeriesRequest{
		Authorization:  actor.Authorization(),
		Schema:         &proto.Key{Key: keystone.Type(evn), Source: actor.VendorApp()},
		From:           nil,
		Until:          nil,
		Interval:       "1 hour",
		SeriesProperty: "library",
	})

	logger.I().Info("ChartTimeSeries", zap.Any("res", res), zap.Error(chartErr))

	return requirements.TestResult{
		Name:  "Chart TimeSeries",
		Error: chartErr,
	}
}
