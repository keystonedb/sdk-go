package query

import (
	"context"
	"fmt"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

func (d *Requirement) readCursorPages(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Query cursor pages, total and projection"}
	cursor := ""
	seen := map[string]bool{}
	cursors := map[string]bool{}
	for page := 0; page < 4; page++ {
		response, err := actor.Query(context.Background(), keystone.Type(models.FileData{}),
			keystone.Select("identifier", "check_key"), keystone.Limit(2), keystone.IncludeTotal(), keystone.Cursor(cursor),
			keystone.WhereEquals("check_key", d.runID), keystone.SortAsc("identifier"))
		if err != nil {
			return result.WithError(err)
		}
		if response.TotalResults == nil || response.GetTotalResults() != 6 {
			return result.WithError(fmt.Errorf("expected explicit total 6, got %v", response.TotalResults))
		}
		if len(response.Entities) != 2 {
			return result.WithError(fmt.Errorf("expected page of 2, got %d", len(response.Entities)))
		}
		for _, entity := range response.Entities {
			file := &models.FileData{}
			if err := keystone.Unmarshal(entity, file); err != nil {
				return result.WithError(err)
			}
			expected := fmt.Sprintf("fd%d", len(seen))
			if file.Identifier != expected || seen[file.Identifier] || file.CheckKey != d.runID {
				return result.WithError(fmt.Errorf("unexpected cursor result %q, expected %q", file.Identifier, expected))
			}
			if file.LineInformation != "" {
				return result.WithError(fmt.Errorf("unselected line_information was returned"))
			}
			seen[file.Identifier] = true
		}
		if !response.HasMore {
			if len(seen) != 6 {
				return result.WithError(fmt.Errorf("expected 6 unique entities, got %d", len(seen)))
			}
			return result
		}
		if response.NextCursor == "" || cursors[response.NextCursor] {
			return result.WithError(fmt.Errorf("missing or repeated next cursor"))
		}
		cursors[response.NextCursor] = true
		cursor = response.NextCursor
	}
	return result.WithError(fmt.Errorf("cursor did not terminate"))
}

func (d *Requirement) readEmptyEntityIDs(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Query explicit empty entity IDs matches nothing"}
	response, err := actor.Query(context.Background(), keystone.Type(models.FileData{}), keystone.IncludeTotal(),
		keystone.WithEntityIDs(nil), keystone.WhereEquals("check_key", d.runID))
	if err != nil {
		return result.WithError(err)
	}
	if len(response.Entities) != 0 || response.HasMore || response.TotalResults == nil || response.GetTotalResults() != 0 {
		return result.WithError(fmt.Errorf("expected no matches, got %v", response))
	}
	return result
}
