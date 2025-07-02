package main

import (
	"github.com/keystonedb/sdk-go/test/requirements"
	"github.com/keystonedb/sdk-go/test/requirements/akv"
	"github.com/keystonedb/sdk-go/test/requirements/child_entities"
	"github.com/keystonedb/sdk-go/test/requirements/cru"
	"github.com/keystonedb/sdk-go/test/requirements/daily"
	"github.com/keystonedb/sdk-go/test/requirements/datatypes"
	"github.com/keystonedb/sdk-go/test/requirements/destroy"
	"github.com/keystonedb/sdk-go/test/requirements/dynamic_entity"
	"github.com/keystonedb/sdk-go/test/requirements/dynamic_properties"
	"github.com/keystonedb/sdk-go/test/requirements/embedded"
	"github.com/keystonedb/sdk-go/test/requirements/events"
	"github.com/keystonedb/sdk-go/test/requirements/exists"
	"github.com/keystonedb/sdk-go/test/requirements/group_count"
	"github.com/keystonedb/sdk-go/test/requirements/hashed_id"
	"github.com/keystonedb/sdk-go/test/requirements/iid"
	"github.com/keystonedb/sdk-go/test/requirements/immutable"
	"github.com/keystonedb/sdk-go/test/requirements/labels"
	"github.com/keystonedb/sdk-go/test/requirements/list"
	"github.com/keystonedb/sdk-go/test/requirements/logging"
	"github.com/keystonedb/sdk-go/test/requirements/lookup"
	"github.com/keystonedb/sdk-go/test/requirements/nested_children"
	"github.com/keystonedb/sdk-go/test/requirements/objects"
	"github.com/keystonedb/sdk-go/test/requirements/pii"
	"github.com/keystonedb/sdk-go/test/requirements/prewrite"
	"github.com/keystonedb/sdk-go/test/requirements/ratelimit"
	"github.com/keystonedb/sdk-go/test/requirements/relationships"
	"github.com/keystonedb/sdk-go/test/requirements/remote"
	"github.com/keystonedb/sdk-go/test/requirements/sensor"
	"github.com/keystonedb/sdk-go/test/requirements/setfalse"
	"github.com/keystonedb/sdk-go/test/requirements/shared_views"
	"github.com/keystonedb/sdk-go/test/requirements/snapshot"
	"github.com/keystonedb/sdk-go/test/requirements/squid"
	"github.com/keystonedb/sdk-go/test/requirements/stats"
	"github.com/keystonedb/sdk-go/test/requirements/status"
	"github.com/keystonedb/sdk-go/test/requirements/stringset"
	"github.com/keystonedb/sdk-go/test/requirements/tasks"
	"github.com/keystonedb/sdk-go/test/requirements/timeseries"
	"github.com/keystonedb/sdk-go/test/requirements/unique_id"
	"github.com/keystonedb/sdk-go/test/requirements/watcher"
)

var reqs []requirements.Requirement

func init() {
	//reqs = append(reqs, &requirements.DummyRequirement{})
	reqs = append(reqs, &akv.Requirement{})
	reqs = append(reqs, &cru.Requirement{})
	reqs = append(reqs, &dynamic_properties.Requirement{})
	reqs = append(reqs, &dynamic_entity.Requirement{})
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
	reqs = append(reqs, &prewrite.Requirement{})
	reqs = append(reqs, &objects.Requirement{})
	reqs = append(reqs, &remote.Requirement{})
	reqs = append(reqs, &nested_children.Requirement{})
	reqs = append(reqs, &embedded.Requirement{})
	reqs = append(reqs, &pii.Requirement{})
	reqs = append(reqs, &ratelimit.Requirement{})
	reqs = append(reqs, &iid.Requirement{})
	reqs = append(reqs, &watcher.Requirement{})
	reqs = append(reqs, &exists.Requirement{})
	reqs = append(reqs, &destroy.Requirement{})
	reqs = append(reqs, &group_count.Requirement{})
	reqs = append(reqs, &hashed_id.Requirement{})
	//reqs = append(reqs, &event_stream.Requirement{})
	reqs = append(reqs, &tasks.Requirement{})
	reqs = append(reqs, &squid.Requirement{})
	reqs = append(reqs, &snapshot.Requirement{})
	reqs = append(reqs, &status.Requirement{})

	if false {
		reqs = []requirements.Requirement{}
		reqs = append(reqs, &status.Requirement{})
		//reqs = append(reqs, &tasks.Requirement{})
	}
}
