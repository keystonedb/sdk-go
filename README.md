# Keystone Go SDK

`Actor.Query` accepts composable options for filters, relationships, projections,
sorting, and cursor pagination:

```go
response, err := actor.Query(ctx, keystone.Type(models.Offer{}),
    keystone.Where("s", "=", "Y"),
    keystone.HasParent("parent-eid"),
    keystone.HasRelation("owns", "related-eid"),
    keystone.Select("prop", "prop2"),
    keystone.SelectSecure("secProp2"),
    keystone.SortAsc("prop"),
    keystone.Limit(11),
    keystone.IncludeTotal(),
)
```

`Select` adds properties to the projection. `SelectSecure` adds properties to both
selection and decryption. Repeated selections merge without duplicates. With no
selection, the query returns entity metadata only.

`HasParent` restricts results to children of an entity. `HasRelation(type, ids...)`
requires an outgoing relationship to any listed entity; omit the IDs to match any
destination. Repeated `HasRelation` calls are ANDed. Use `RelationOf` for incoming
relationships, or `RelationOf`/`RelationTo` to specify the vendor and app.

Existing `FindOption` helpers also work, including `Where`, `And`, `Or`, labels,
entity IDs, state filters, and sorting. `WithEntityIDs(nil)` explicitly matches
nothing; omit the option to leave entity IDs unrestricted. Child schemas require
a parent ID.

The response retains `Entities`, aggregate `Rows`, `NextCursor`, `HasMore`, optional
`TotalResults`, and `Execution`. Add `keystone.Cursor(response.NextCursor)` to the
same options to fetch the next page. Query uses cursor pagination: `Limit(size)`
and `Limit(size, 0)` are supported; nonzero page numbers are rejected.

For aggregate queries and other advanced protobuf settings, use
`Actor.QueryWithRequest`. The request is copied before options and actor
authorization are applied:

```go
response, err := actor.QueryWithRequest(ctx, keystone.Type(models.Offer{}), &proto.QueryRequest{
    GroupBy: []string{"config_type"},
    Aggregates: []*proto.QueryAggregate{
        {Function: proto.QueryAggregate_COUNT, As: "count"},
    },
}, keystone.Where("state", "=", 1), keystone.Limit(25))
```

`GroupBy`, `Aggregates`, and `Having` return group keys and aggregate values in
`Rows`. Existing calls using `actor.Query(ctx, entityType, request, options...)`
should use `actor.QueryWithRequest(ctx, entityType, request, options...)`.

The `test/requirements/query` suite reuses existing Find, GroupCount, and QueryIndex
requirement scenarios and adds cursor, projection, total, HAVING, and empty-ID
coverage. The requirements runner needs a Keystone server at `127.0.0.1:50051`;
`go test ./...` runs the SDK unit tests and compiles the requirements.
