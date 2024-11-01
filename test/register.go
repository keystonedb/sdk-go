package main

import (
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/keystonedb/sdk-go/test/requirements/child_entities"
	"github.com/keystonedb/sdk-go/test/requirements/cru"
	"github.com/keystonedb/sdk-go/test/requirements/daily"
	"github.com/keystonedb/sdk-go/test/requirements/datatypes"
	"github.com/keystonedb/sdk-go/test/requirements/dynamic_properties"
	"github.com/keystonedb/sdk-go/test/requirements/events"
	"github.com/keystonedb/sdk-go/test/requirements/immutable"
	"github.com/keystonedb/sdk-go/test/requirements/labels"
	"github.com/keystonedb/sdk-go/test/requirements/list"
	"github.com/keystonedb/sdk-go/test/requirements/logging"
	"github.com/keystonedb/sdk-go/test/requirements/lookup"
	"github.com/keystonedb/sdk-go/test/requirements/ratelimit"
	"github.com/keystonedb/sdk-go/test/requirements/relationships"
	"github.com/keystonedb/sdk-go/test/requirements/sensor"
	"github.com/keystonedb/sdk-go/test/requirements/setfalse"
	"github.com/keystonedb/sdk-go/test/requirements/shared_views"
	"github.com/keystonedb/sdk-go/test/requirements/stats"
	"github.com/keystonedb/sdk-go/test/requirements/stringset"
	"github.com/keystonedb/sdk-go/test/requirements/timeseries"
	"github.com/keystonedb/sdk-go/test/requirements/unique_id"
)

var reqs []requirements.Requirement

func init() {
	//reqs = append(reqs, &requirements.DummyRequirement{})
	reqs = append(reqs, &cru.Requirement{})
	reqs = append(reqs, &dynamic_properties.Requirement{})
	reqs = append(reqs, &unique_id.Requirement{})
	reqs = append(reqs, &sensor.Requirement{})
	reqs = append(reqs, &labels.Requirement{})
	reqs = append(reqs, &logging.Requirement{})
	reqs = append(reqs, &child_entities.Requirement{})
	reqs = append(reqs, &events.Requirement{})
	reqs = append(reqs, &daily.Requirement{})
	reqs = append(reqs, &stats.Requirement{})
	reqs = append(reqs, &immutable.Requirement{})
	reqs = append(reqs, &lookup.Requirement{})
	reqs = append(reqs, &relationships.Requirement{})
	reqs = append(reqs, &list.Requirement{})
	reqs = append(reqs, &datatypes.Requirement{})
	reqs = append(reqs, &timeseries.Requirement{})
	reqs = append(reqs, &setfalse.Requirement{})
	reqs = append(reqs, &shared_views.Requirement{})
	reqs = append(reqs, &stringset.Requirement{})
	reqs = append(reqs, &ratelimit.Requirement{})

	if true {
		reqs = []requirements.Requirement{}
		reqs = append(reqs, &ratelimit.Requirement{})
	}
}
