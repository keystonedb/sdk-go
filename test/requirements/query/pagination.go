package query

import (
	"context"
	"fmt"
	"time"

	"github.com/keystonedb/sdk-go/keystone"
	"github.com/keystonedb/sdk-go/test/models"
	"github.com/keystonedb/sdk-go/test/requirements"
)

func (d *Requirement) readSavedQueryPages(actor *keystone.Actor) requirements.TestResult {
	result := requirements.TestResult{Name: "Query returns first saved page, then pages by result token"}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Saving the query must return the first page immediately, alongside the
	// token and cursor used to read subsequent pages without the original query.
	response, err := actor.Query(ctx, keystone.Type(models.FileData{}),
		keystone.WhereEquals("check_key", d.runID),
		keystone.Select("identifier", "check_key"), keystone.SortAsc("identifier"),
		keystone.Limit(4), keystone.SaveResultSet(), keystone.IncludeTotal(), keystone.Explain())
	if err != nil {
		return result.WithError(err)
	}
	if len(response.Entities) != 4 || !response.HasMore || response.NextCursor == "" || response.ResultSet == "" || response.ResultSetExpires == nil {
		return result.WithError(fmt.Errorf("initial saved query must return four entities, token, expiry and next cursor: %v", response))
	}
	token, expires := response.ResultSet, response.ResultSetExpires.AsTime()
	if response.ResultSetExpires.CheckValid() != nil || !expires.After(time.Now()) {
		return result.WithError(fmt.Errorf("invalid saved query expiry: %v", response.ResultSetExpires))
	}
	seen := map[string]bool{}
	cursors := map[string]bool{}
	for page, size := range []int{4, 2} {
		if len(response.Entities) != size || len(response.Rows) != 0 || response.TotalResults == nil || response.GetTotalResults() != 6 || response.ResultSet != token || response.ResultSetExpires == nil || !response.ResultSetExpires.AsTime().Equal(expires) {
			return result.WithError(fmt.Errorf("unexpected saved query page %d: %v", page, response))
		}
		if page > 0 && response.GetExecution().GetStrategy() != "result_set" {
			return result.WithError(fmt.Errorf("continuation did not read the saved result set: %v", response.Execution))
		}
		for _, entity := range response.Entities {
			file := &models.FileData{}
			if err := keystone.Unmarshal(entity, file); err != nil {
				return result.WithError(err)
			}
			expected := fmt.Sprintf("fd%d", len(seen))
			id := file.GetKeystoneID().String()
			if id == "" || seen[id] || file.Identifier != expected || file.CheckKey != d.runID || file.LineInformation != "" {
				return result.WithError(fmt.Errorf("unexpected saved page entity %q, expected %q (ID %q)", file.Identifier, expected, id))
			}
			seen[id] = true
		}
		if page == 1 {
			if response.HasMore || response.NextCursor != "" || len(seen) != 6 {
				return result.WithError(fmt.Errorf("final partial page must exhaust six unique results: %v", response))
			}
			return result
		}
		if !response.HasMore || response.NextCursor == "" || cursors[response.NextCursor] {
			return result.WithError(fmt.Errorf("missing or repeated saved query cursor"))
		}
		cursors[response.NextCursor] = true
		// Only the saved token and paging controls accompany subsequent reads.
		response, err = actor.Query(ctx, "", keystone.ResultSet(token),
			keystone.Cursor(response.NextCursor), keystone.Limit(4), keystone.IncludeTotal(), keystone.Explain())
		if err != nil {
			return result.WithError(err)
		}
	}
	return result.WithError(fmt.Errorf("saved query pagination did not terminate"))
}

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
