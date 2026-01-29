package keystone

import (
	"testing"

	"github.com/keystonedb/sdk-go/proto"
)

// TestWhereEquals tests the WhereEquals find option
func TestFindWhereEquals(t *testing.T) {
	opt := WhereEquals("name", "John")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "name" {
		t.Errorf("expected property 'name', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_Equal {
		t.Errorf("expected Operator_Equal, got %v", filter.Operator)
	}
}

// TestWhereNotEquals tests the WhereNotEquals find option
func TestFindWhereNotEquals(t *testing.T) {
	opt := WhereNotEquals("status", "inactive")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "status" {
		t.Errorf("expected property 'status', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_NotEqual {
		t.Errorf("expected Operator_NotEqual, got %v", filter.Operator)
	}
}

// TestWhereGreaterThan tests the WhereGreaterThan find option
func TestFindWhereGreaterThan(t *testing.T) {
	opt := WhereGreaterThan("age", 18)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "age" {
		t.Errorf("expected property 'age', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_GreaterThan {
		t.Errorf("expected Operator_GreaterThan, got %v", filter.Operator)
	}
}

// TestWhereGreaterThanOrEquals tests the WhereGreaterThanOrEquals find option
func TestFindWhereGreaterThanOrEquals(t *testing.T) {
	opt := WhereGreaterThanOrEquals("score", 100)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "score" {
		t.Errorf("expected property 'score', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_GreaterThanOrEqual {
		t.Errorf("expected Operator_GreaterThanOrEqual, got %v", filter.Operator)
	}
}

// TestWhereLessThan tests the WhereLessThan find option
func TestFindWhereLessThan(t *testing.T) {
	opt := WhereLessThan("price", 50)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "price" {
		t.Errorf("expected property 'price', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_LessThan {
		t.Errorf("expected Operator_LessThan, got %v", filter.Operator)
	}
}

// TestWhereLessThanOrEquals tests the WhereLessThanOrEquals find option
func TestFindWhereLessThanOrEquals(t *testing.T) {
	opt := WhereLessThanOrEquals("quantity", 10)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "quantity" {
		t.Errorf("expected property 'quantity', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_LessThanOrEqual {
		t.Errorf("expected Operator_LessThanOrEqual, got %v", filter.Operator)
	}
}

// TestWhereContains tests the WhereContains find option
func TestFindWhereContains(t *testing.T) {
	opt := WhereContains("description", "test")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "description" {
		t.Errorf("expected property 'description', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_Contains {
		t.Errorf("expected Operator_Contains, got %v", filter.Operator)
	}
}

// TestWhereNotContains tests the WhereNotContains find option
func TestFindWhereNotContains(t *testing.T) {
	opt := WhereNotContains("tags", "spam")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "tags" {
		t.Errorf("expected property 'tags', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_NotContains {
		t.Errorf("expected Operator_NotContains, got %v", filter.Operator)
	}
}

// TestWhereStartsWith tests the WhereStartsWith find option
func TestFindWhereStartsWith(t *testing.T) {
	opt := WhereStartsWith("email", "admin")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "email" {
		t.Errorf("expected property 'email', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_StartsWith {
		t.Errorf("expected Operator_StartsWith, got %v", filter.Operator)
	}
}

// TestWhereEndsWith tests the WhereEndsWith find option
func TestFindWhereEndsWith(t *testing.T) {
	opt := WhereEndsWith("filename", ".pdf")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "filename" {
		t.Errorf("expected property 'filename', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_EndsWith {
		t.Errorf("expected Operator_EndsWith, got %v", filter.Operator)
	}
}

// TestWhereIn tests the WhereIn find option
func TestFindWhereIn(t *testing.T) {
	opt := WhereIn("status", "active", "pending", "review")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "status" {
		t.Errorf("expected property 'status', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_In {
		t.Errorf("expected Operator_In, got %v", filter.Operator)
	}
	if len(filter.Values) != 3 {
		t.Errorf("expected 3 values, got %d", len(filter.Values))
	}
}

// TestWhereNotIn tests the WhereNotIn find option
func TestFindWhereNotIn(t *testing.T) {
	opt := WhereNotIn("type", "deleted", "archived")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "type" {
		t.Errorf("expected property 'type', got '%s'", filter.Property)
	}
	// WhereNotIn uses NotEqual operator internally
	if filter.Operator != proto.Operator_NotEqual {
		t.Errorf("expected Operator_NotEqual, got %v", filter.Operator)
	}
}

// TestWhereBetween tests the WhereBetween find option
func TestFindWhereBetween(t *testing.T) {
	opt := WhereBetween("price", 10, 100)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "price" {
		t.Errorf("expected property 'price', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_Between {
		t.Errorf("expected Operator_Between, got %v", filter.Operator)
	}
	if len(filter.Values) != 2 {
		t.Errorf("expected 2 values, got %d", len(filter.Values))
	}
}

// TestIsNull tests the IsNull find option
func TestFindIsNull(t *testing.T) {
	opt := IsNull("deleted_at")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "deleted_at" {
		t.Errorf("expected property 'deleted_at', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_IsNull {
		t.Errorf("expected Operator_IsNull, got %v", filter.Operator)
	}
}

// TestIsNotNull tests the IsNotNull find option
func TestFindIsNotNull(t *testing.T) {
	opt := IsNotNull("verified_at")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Property != "verified_at" {
		t.Errorf("expected property 'verified_at', got '%s'", filter.Property)
	}
	if filter.Operator != proto.Operator_IsNotNull {
		t.Errorf("expected Operator_IsNotNull, got %v", filter.Operator)
	}
}

// TestWhere tests the Where function with various operators
func TestFindWhere(t *testing.T) {
	tests := []struct {
		name             string
		operator         string
		expectedOperator proto.Operator
		values           []any
	}{
		{"eq", "eq", proto.Operator_Equal, []any{"value"}},
		{"=", "=", proto.Operator_Equal, []any{"value"}},
		{"neq", "neq", proto.Operator_NotEqual, []any{"value"}},
		{"!=", "!=", proto.Operator_NotEqual, []any{"value"}},
		{"gt", "gt", proto.Operator_GreaterThan, []any{10}},
		{">", ">", proto.Operator_GreaterThan, []any{10}},
		{"gte", "gte", proto.Operator_GreaterThanOrEqual, []any{10}},
		{">=", ">=", proto.Operator_GreaterThanOrEqual, []any{10}},
		{"lt", "lt", proto.Operator_LessThan, []any{10}},
		{"<", "<", proto.Operator_LessThan, []any{10}},
		{"lte", "lte", proto.Operator_LessThanOrEqual, []any{10}},
		{"<=", "<=", proto.Operator_LessThanOrEqual, []any{10}},
		{"contains", "contains", proto.Operator_Contains, []any{"text"}},
		{"c", "c", proto.Operator_Contains, []any{"text"}},
		{"notcontains", "notcontains", proto.Operator_NotContains, []any{"text"}},
		{"nc", "nc", proto.Operator_NotContains, []any{"text"}},
		{"startswith", "startswith", proto.Operator_StartsWith, []any{"prefix"}},
		{"sw", "sw", proto.Operator_StartsWith, []any{"prefix"}},
		{"endswith", "endswith", proto.Operator_EndsWith, []any{"suffix"}},
		{"ew", "ew", proto.Operator_EndsWith, []any{"suffix"}},
		{"in", "in", proto.Operator_In, []any{"a", "b", "c"}},
		{"notin", "notin", proto.Operator_NotEqual, []any{"a", "b"}},
		{"between", "between", proto.Operator_Between, []any{1, 100}},
		{"btw", "btw", proto.Operator_Between, []any{1, 100}},
		{"><", "><", proto.Operator_Between, []any{1, 100}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := Where("field", tt.operator, tt.values...)
			if opt == nil {
				t.Fatalf("Where returned nil for operator %s", tt.operator)
			}

			config := &filterRequest{}
			opt.Apply(config)

			if len(config.Filters) != 1 {
				t.Fatalf("expected 1 filter, got %d", len(config.Filters))
			}

			filter := config.Filters[0]
			if filter.Operator != tt.expectedOperator {
				t.Errorf("expected operator %v, got %v", tt.expectedOperator, filter.Operator)
			}
		})
	}
}

// TestWhereReturnsNilForEmptyValues tests that Where returns nil when no values are provided
func TestFindWhereReturnsNilForEmptyValues(t *testing.T) {
	opt := Where("field", "eq")
	if opt != nil {
		t.Error("expected nil for Where with no values")
	}
}

// TestWhereReturnsNilForUnknownOperator tests that Where returns nil for unknown operators
func TestFindWhereReturnsNilForUnknownOperator(t *testing.T) {
	opt := Where("field", "unknown", "value")
	if opt != nil {
		t.Error("expected nil for unknown operator")
	}
}

// TestWhereBetweenRequiresTwoValues tests that Where with between operator requires two values
func TestFindWhereBetweenRequiresTwoValues(t *testing.T) {
	opt := Where("field", "between", 10)
	if opt != nil {
		t.Error("expected nil for between with only one value")
	}
}

// TestOr tests the Or composite filter
func TestFindOr(t *testing.T) {
	opt := Or(
		WhereEquals("status", "active"),
		WhereEquals("status", "pending"),
	)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if !filter.Or {
		t.Error("expected Or flag to be true")
	}
	if len(filter.Nested) != 2 {
		t.Errorf("expected 2 nested filters, got %d", len(filter.Nested))
	}
}

// TestAnd tests the And composite filter
func TestFindAnd(t *testing.T) {
	opt := And(
		WhereEquals("status", "active"),
		WhereGreaterThan("age", 18),
	)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(config.Filters))
	}

	filter := config.Filters[0]
	if filter.Or {
		t.Error("expected Or flag to be false")
	}
	if len(filter.Nested) != 2 {
		t.Errorf("expected 2 nested filters, got %d", len(filter.Nested))
	}
}

// TestSortBy tests the SortBy find option
func TestFindSortBy(t *testing.T) {
	opt := SortBy("created_at", true)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.sortBy) != 1 {
		t.Fatalf("expected 1 sort, got %d", len(config.sortBy))
	}

	sort := config.sortBy[0]
	if sort.Property != "created_at" {
		t.Errorf("expected property 'created_at', got '%s'", sort.Property)
	}
	if !sort.Descending {
		t.Error("expected Descending to be true")
	}
}

// TestSortDesc tests the SortDesc find option
func TestFindSortDesc(t *testing.T) {
	opt := SortDesc("updated_at")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.sortBy) != 1 {
		t.Fatalf("expected 1 sort, got %d", len(config.sortBy))
	}

	sort := config.sortBy[0]
	if sort.Property != "updated_at" {
		t.Errorf("expected property 'updated_at', got '%s'", sort.Property)
	}
	if !sort.Descending {
		t.Error("expected Descending to be true")
	}
}

// TestSortAsc tests the SortAsc find option
func TestFindSortAsc(t *testing.T) {
	opt := SortAsc("name")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.sortBy) != 1 {
		t.Fatalf("expected 1 sort, got %d", len(config.sortBy))
	}

	sort := config.sortBy[0]
	if sort.Property != "name" {
		t.Errorf("expected property 'name', got '%s'", sort.Property)
	}
	if sort.Descending {
		t.Error("expected Descending to be false")
	}
}

// TestSortByNullFirst tests the SortByNullFirst find option
func TestFindSortByNullFirst(t *testing.T) {
	opt := SortByNullFirst("optional_field", false)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.sortBy) != 1 {
		t.Fatalf("expected 1 sort, got %d", len(config.sortBy))
	}

	sort := config.sortBy[0]
	if sort.Property != "optional_field" {
		t.Errorf("expected property 'optional_field', got '%s'", sort.Property)
	}
	if sort.Descending {
		t.Error("expected Descending to be false")
	}
	if !sort.NullsFirst {
		t.Error("expected NullsFirst to be true")
	}
}

// TestWithEntityIDs tests the WithEntityIDs find option
func TestFindWithEntityIDs(t *testing.T) {
	ids := []string{"id1", "id2", "id3"}
	opt := WithEntityIDs(ids)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.EntityIds) != 3 {
		t.Fatalf("expected 3 entity IDs, got %d", len(config.EntityIds))
	}
	if config.EntityIds[0] != "id1" {
		t.Errorf("expected first ID 'id1', got '%s'", config.EntityIds[0])
	}
	if config.EntityIds[1] != "id2" {
		t.Errorf("expected second ID 'id2', got '%s'", config.EntityIds[1])
	}
	if config.EntityIds[2] != "id3" {
		t.Errorf("expected third ID 'id3', got '%s'", config.EntityIds[2])
	}
}

// TestWithLabel tests the WithLabel find option
func TestFindWithLabel(t *testing.T) {
	opt := WithLabel("category", "featured")
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Labels) != 1 {
		t.Fatalf("expected 1 label, got %d", len(config.Labels))
	}

	label := config.Labels[0]
	if label.Name != "category" {
		t.Errorf("expected label name 'category', got '%s'", label.Name)
	}
	if label.Value != "featured" {
		t.Errorf("expected label value 'featured', got '%s'", label.Value)
	}
}

// TestWithLabelMultiple tests multiple WithLabel options
func TestFindWithLabelMultiple(t *testing.T) {
	opt1 := WithLabel("category", "featured")
	opt2 := WithLabel("type", "product")
	config := &filterRequest{}

	opt1.Apply(config)
	opt2.Apply(config)

	if len(config.Labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(config.Labels))
	}
}

// TestLimit tests the Limit find option
func TestFindLimit(t *testing.T) {
	opt := Limit(25, 2)
	config := &filterRequest{}

	opt.Apply(config)

	if config.PerPage != 25 {
		t.Errorf("expected PerPage 25, got %d", config.PerPage)
	}
	if config.PageNumber != 2 {
		t.Errorf("expected PageNumber 2, got %d", config.PageNumber)
	}
}

// TestChildOf tests the ChildOf find option
func TestFindChildOf(t *testing.T) {
	opt := ChildOf("parent-entity-123")
	config := &filterRequest{}

	opt.Apply(config)

	if config.ParentEntityID != "parent-entity-123" {
		t.Errorf("expected ParentEntityID 'parent-entity-123', got '%s'", config.ParentEntityID)
	}
}

// TestRelationOf tests the RelationOf find option
func TestFindRelationOf(t *testing.T) {
	opt := RelationOf("entity-123", "owns", "vendor1", "app1")
	config := &filterRequest{}

	opt.Apply(config)

	if config.RelationOf == nil {
		t.Fatal("expected RelationOf to be set")
	}
	if config.RelationOf.SourceId != "entity-123" {
		t.Errorf("expected SourceId 'entity-123', got '%s'", config.RelationOf.SourceId)
	}
	if config.RelationOf.Relationship.Key != "owns" {
		t.Errorf("expected Relationship.Key 'owns', got '%s'", config.RelationOf.Relationship.Key)
	}
	if config.RelationOf.Relationship.Source.VendorId != "vendor1" {
		t.Errorf("expected VendorId 'vendor1', got '%s'", config.RelationOf.Relationship.Source.VendorId)
	}
	if config.RelationOf.Relationship.Source.AppId != "app1" {
		t.Errorf("expected AppId 'app1', got '%s'", config.RelationOf.Relationship.Source.AppId)
	}
}

// TestRelationTo tests the RelationTo find option
func TestFindRelationTo(t *testing.T) {
	entityID := ID("dest-entity-456")
	opt := RelationTo(entityID, "belongs_to", "vendor2", "app2")
	config := &filterRequest{}

	opt.Apply(config)

	if config.RelationOf == nil {
		t.Fatal("expected RelationOf to be set")
	}
	if config.RelationOf.DestinationId != "dest-entity-456" {
		t.Errorf("expected DestinationId 'dest-entity-456', got '%s'", config.RelationOf.DestinationId)
	}
	if config.RelationOf.Relationship.Key != "belongs_to" {
		t.Errorf("expected Relationship.Key 'belongs_to', got '%s'", config.RelationOf.Relationship.Key)
	}
}

// TestRelationOfSibling tests the RelationOfSibling find option
func TestFindRelationOfSibling(t *testing.T) {
	opt := RelationOfSibling("entity-789", "sibling_of")
	config := &filterRequest{}

	opt.Apply(config)

	if config.RelationOf == nil {
		t.Fatal("expected RelationOf to be set")
	}
	if config.RelationOf.SourceId != "entity-789" {
		t.Errorf("expected SourceId 'entity-789', got '%s'", config.RelationOf.SourceId)
	}
	if config.RelationOf.Relationship.Key != "sibling_of" {
		t.Errorf("expected Relationship.Key 'sibling_of', got '%s'", config.RelationOf.Relationship.Key)
	}
	// Sibling relations should have empty vendor/app
	if config.RelationOf.Relationship.Source.VendorId != "" {
		t.Errorf("expected empty VendorId for sibling, got '%s'", config.RelationOf.Relationship.Source.VendorId)
	}
	if config.RelationOf.Relationship.Source.AppId != "" {
		t.Errorf("expected empty AppId for sibling, got '%s'", config.RelationOf.Relationship.Source.AppId)
	}
}

// TestRelationToSibling tests the RelationToSibling find option
func TestFindRelationToSibling(t *testing.T) {
	entityID := ID("dest-entity-000")
	opt := RelationToSibling(entityID, "linked_to")
	config := &filterRequest{}

	opt.Apply(config)

	if config.RelationOf == nil {
		t.Fatal("expected RelationOf to be set")
	}
	if config.RelationOf.DestinationId != "dest-entity-000" {
		t.Errorf("expected DestinationId 'dest-entity-000', got '%s'", config.RelationOf.DestinationId)
	}
	// Sibling relations should have empty vendor/app
	if config.RelationOf.Relationship.Source.VendorId != "" {
		t.Errorf("expected empty VendorId for sibling, got '%s'", config.RelationOf.Relationship.Source.VendorId)
	}
}

// TestMultipleFindOptions tests applying multiple find options
func TestFindMultipleOptions(t *testing.T) {
	config := &filterRequest{}

	// Apply multiple options
	WhereEquals("status", "active").Apply(config)
	WhereGreaterThan("age", 18).Apply(config)
	SortDesc("created_at").Apply(config)
	Limit(10, 1).Apply(config)
	WithLabel("category", "premium").Apply(config)

	// Verify filters
	if len(config.Filters) != 2 {
		t.Errorf("expected 2 filters, got %d", len(config.Filters))
	}

	// Verify sort
	if len(config.sortBy) != 1 {
		t.Errorf("expected 1 sort, got %d", len(config.sortBy))
	}

	// Verify pagination
	if config.PerPage != 10 {
		t.Errorf("expected PerPage 10, got %d", config.PerPage)
	}
	if config.PageNumber != 1 {
		t.Errorf("expected PageNumber 1, got %d", config.PageNumber)
	}

	// Verify labels
	if len(config.Labels) != 1 {
		t.Errorf("expected 1 label, got %d", len(config.Labels))
	}
}

// TestNestedOrAndFilters tests nested Or and And filters
func TestFindNestedOrAndFilters(t *testing.T) {
	opt := Or(
		And(
			WhereEquals("status", "active"),
			WhereGreaterThan("score", 80),
		),
		And(
			WhereEquals("status", "vip"),
			WhereGreaterThan("score", 50),
		),
	)
	config := &filterRequest{}

	opt.Apply(config)

	if len(config.Filters) != 1 {
		t.Fatalf("expected 1 top-level filter, got %d", len(config.Filters))
	}

	topFilter := config.Filters[0]
	if !topFilter.Or {
		t.Error("expected top-level filter to be Or")
	}
	if len(topFilter.Nested) != 2 {
		t.Errorf("expected 2 nested And filters, got %d", len(topFilter.Nested))
	}

	// Each nested And filter should have 2 nested filters
	for i, nested := range topFilter.Nested {
		if len(nested.Nested) != 2 {
			t.Errorf("expected nested And filter %d to have 2 filters, got %d", i, len(nested.Nested))
		}
	}
}

// TestValueConversions tests value conversion helper functions
func TestFindValueConversions(t *testing.T) {
	// Test valueFromString
	strVal := valueFromString("test")
	if strVal.Text != "test" {
		t.Errorf("expected Text 'test', got '%s'", strVal.Text)
	}

	// Test valueFromInt
	intVal := valueFromInt(42)
	if intVal.Int != 42 {
		t.Errorf("expected Int 42, got %d", intVal.Int)
	}

	// Test nilValue
	nullVal := nilValue()
	if !nullVal.IsNull {
		t.Error("expected IsNull to be true")
	}

	// Test valueFromAny with nil
	nilAnyVal := valueFromAny(nil)
	if nilAnyVal != nil {
		t.Error("expected nil for valueFromAny(nil)")
	}

	// Test valuesFromAny with mixed values including nil
	vals := valuesFromAny("test", nil, 42)
	if len(vals) != 3 {
		t.Fatalf("expected 3 values, got %d", len(vals))
	}
	// Second value should be null
	if !vals[1].IsNull {
		t.Error("expected second value to be null")
	}
}

// TestMultipleSortOptions tests multiple sort options
func TestFindMultipleSortOptions(t *testing.T) {
	config := &filterRequest{}

	SortDesc("created_at").Apply(config)
	SortAsc("name").Apply(config)

	if len(config.sortBy) != 2 {
		t.Fatalf("expected 2 sorts, got %d", len(config.sortBy))
	}

	if config.sortBy[0].Property != "created_at" || !config.sortBy[0].Descending {
		t.Error("first sort should be created_at descending")
	}
	if config.sortBy[1].Property != "name" || config.sortBy[1].Descending {
		t.Error("second sort should be name ascending")
	}
}
