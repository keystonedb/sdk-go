# Keystone Go SDK

## Ordered feeds

Use `actor.Feed(type, key)` for channel messages, replies or user activity.
Feeds use Cassandra directly and do not require SQL indexing or schema registration.
Public keys can be existing entity IDs, k4id strings or ordinary names. For
example, `actor.Feed("channel-messages", k4id.New().String())` selects a new feed.
The server derives its internal 16-byte ID from the full actor scope, type and
key, and stores UTC hours in four bytes. The SDK keeps sending the original
type/key values; positions remain independent UUID v1 ordering keys.

```go
messages := actor.Feed("channel-messages", channelID)
entry := &keystone.FeedEntry{
    EntityID:   message.GetKeystoneID(),
    EntityType: keystone.Type(message),
}
if err := messages.Put(ctx, entry); err != nil {
    return err
}

page, err := messages.Read(ctx, keystone.FeedReadOptions{Limit: 50})
if err != nil {
    return err
}
for {
    for _, entry := range page.Entries() {
        fmt.Println(entry.EntityID, entry.Position)
    }
    if !page.HasMore() {
        break
    }
    if err := page.Next(ctx); err != nil {
        return err
    }
}
```

`Put` fills in `entry.Position` before sending. Keep the entry and retry `Put`
with that same entry on failure: its position is preserved so retries do not
create duplicate membership rows. To update or remove an existing membership,
use its full position; `messages.Remove(ctx, entry.Position)` removes the feed
entry without deleting the referenced entity. Feed writes and entity mutations
are separate operations.

`FeedReadOptions{}` reads the newest 50 entries. Set `Ascending: true` for oldest
first. Set `Position` to read strictly before an entry, or strictly after it when
ascending. Set `Cursor` to resume a saved `page.NextCursor()` with the same feed,
direction and actor identity; do not combine it with `Position`. Limits cannot
exceed 1,000 and payload-heavy pages may be shorter. Positions order by UUID v1
timestamp, rather than commit order; reads are not snapshots.

`Next` replaces the current page, returns `io.EOF` when exhausted and preserves
the current page and cursor on failure. Always check `HasMore`, even for an empty
page: sparse feeds can require another bounded read. The pager retains the
original connection, authorization and options. Use entries and pagers sequentially.

`FeedEntry` holds an entity reference, an opaque `Data []byte` payload (up to
16 KiB), or both. References are not automatically loaded. Use `actor.GetByID`
to retrieve the entity through the primary store. For payload-only activity:

```go
activity := actor.Feed("user-activity", userID)
err := activity.Put(ctx, &keystone.FeedEntry{Data: []byte(`{"action":"mentioned"}`)})
```

Thread replies use the same API with `actor.Feed("message-replies", messageID)`.
For calendar navigation or an activity scrubber, request a summary:

```go
summary, err := messages.Summary(ctx, keystone.FeedSummaryOptions{
    From: start, Until: end,
    Interval: keystone.FeedDaily,
    Timezone: "Europe/London",
})
if err != nil {
    return err
}
for _, day := range summary.Buckets {
    fmt.Println(day.Start, day.Count)
}
// After selecting a bucket, jump into the feed:
page, err := messages.Read(ctx, keystone.FeedReadOptions{
    Cursor: summary.Buckets[selected].Cursor,
    Limit: 50,
})
```

Intervals are `FeedHourly`, `FeedDaily` (default) and `FeedMonthly`. Ranges are
inclusive of `From` and exclusive of `Until`, with both timestamps aligned to a
minute. Daily/monthly ranges can cover 366 days, hourly ranges 31 days. Timezone
defaults to UTC; fractional-hour offsets and DST are handled by the server.
Bucket timestamps are absolute instants: render them in the requested timezone.
Possible activity is retained; jump boundaries are clipped to the requested range.

`summary.Approximate` is true for conservative activity outlines. Counts estimate
activity density; retries, updates and removals can overcount. Every period
overlapping a recorded hour stays visible even if its counter is missing or zero,
so navigation does not wait for a background projection. Empty periods can remain
visible after removal, and fractional timezone boundaries can conservatively
highlight adjacent periods. Do not use these counts as exact row totals.

Use the bucket's `Cursor` to navigate by time:
set `Ascending: true` in both Summary and Read to start at the period's beginning,
or omit it to start before the period's end. Jumps include boundary entries where
applicable and continue through the feed. Normal cursor expiry and scope rules
apply. Summary reads use the directory and counters, without message scans or SQL.

The old exact-projection fields (`PendingHours`, `FirstPosition`, `LastPosition`,
`UpdatedAt`) remain deprecated for source compatibility. Current activity
outlines are checked immediately; applications do not need to poll a projector.

The feed requirements use short k4id keys and exercise scope isolation, ambiguous
namespace boundaries, Unicode keys, signed hour ordering across the Unix epoch,
and calendar jumps in both directions. Activity checks permit density drift and
extra historical markers while failing immediately if an active period is absent.

For advanced wire settings, the generated `Connection.FeedPut`, `FeedRead` and
`FeedRemove` and `FeedSummary` RPC methods remain available.

## Entity queries

Use a typed collection to keep loaded entities and query pagination together:

```go
customers := keystone.Collection(Customer{}, actor)
if err := customers.Query(ctx,
    keystone.Select("name", "email"),
    keystone.SortAsc("name"),
    keystone.Limit(25),
    keystone.IncludeTotal(),
); err != nil {
    return err
}

for customer := range customers.All() {
    fmt.Println(customer.Name)
}

if customers.HasMore() {
    if err := customers.Next(ctx); err != nil {
        return err
    }
}
```

`All()` supports Go's range syntax and iterates over loaded entities without
fetching pages. `Items()` exposes the typed slice. `Next` appends another page
using the saved query; it returns `io.EOF` when exhausted. A new successful `Query`
replaces the contents. Failed queries or decoding leave the previous contents and
cursor intact. `Response()` retains the latest page's raw response, including
optional totals and execution details. Collections require sequential access.

The actor is bound at construction and reused by `Query` and `Next`.
Use `customers.Latest(ctx, 10, keystone.Select("name", "email"))` to replace the
contents with up to ten entities ordered by creation time, newest first. The size
must be positive. Optional query options add projections and filters; without a
selection, only metadata is loaded. `Latest` starts at the first page and its size
overrides any `Limit` option. Use `Next(ctx)` to append older results.

The model example supplies the type only; its values are not filters or defaults.
Struct and pointer models are supported: `Collection(Customer{}, actor)` holds
`[]Customer`, while `Collection((*Customer)(nil), actor)` holds `[]*Customer`. Range
variables for struct models are copies; use pointer models to edit loaded
entities during iteration. As with `Actor.Query`, select properties explicitly;
queries with no selection load metadata only.

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

For advanced protobuf settings, use `Actor.QueryWithRequest`. The request is
copied before options and actor authorization are applied:

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

Aggregate collections use SDK-provided rows, so no summary struct is needed:

```go
offers := keystone.Collection(models.Offer{}, actor)
summary, err := offers.Aggregate(ctx,
    keystone.GroupBy("is_global"),
    keystone.Count("count"),
    keystone.Having(keystone.WhereGreaterThan("count", 1)),
    keystone.SortDesc("count"),
    keystone.Limit(25),
    keystone.SaveResultSet(),
    keystone.IncludeTotal(),
)
if err != nil {
    return err
}
for summary.HasMore() {
    if err := summary.Next(ctx); err != nil {
        return err
    }
}
for row := range summary.All() {
    fmt.Println(row.Value("is_global").GetBool(), row.Value("count").GetInt())
}
```

`Aggregate` uses the entity collection's model and actor. It returns a separate
`AggregateResults` collection and leaves loaded entities unchanged. Supply query
filters explicitly; it does not inherit a previous entity query's filters.
`Items()` returns loaded `AggregateRow` values, and `All()` iterates them.
`Next()` appends another page; `Query()` replaces the rows on success. Errors
retain previous rows and cursor state. Both live and saved result pagination are
supported. `Response()` provides the latest page's totals and execution details.

Each row is a map of group keys and aggregate aliases to protobuf values.
For single-row results, use `summary.IntValue("count")`,
`summary.FloatValue("average")`, `summary.BoolValue("flag")`, or
`summary.TextValue("category")` directly. `StringValue` is an alias for `TextValue`.
These accessors read the first loaded row and return zero values for missing
rows or aliases. They read the corresponding protobuf field without conversion.
`Value(alias)` reads the first loaded row and returns nil if no row or alias exists.
`row.Value(alias)` returns nil for a missing alias; `value.IsNull` distinguishes
SQL NULL from zero. Protobuf getters such as `GetInt`, `GetFloat`, and `GetText`
are safe on missing values. Rows and response metadata should be treated as
read-only. Collections require sequential access.

Use `keystone.AggregateCollection(entityType, actor)` to construct a reusable
aggregate collection directly, then call `Query`. A saved aggregate token can be
reopened with `summary.Query(ctx, keystone.ResultSet(token), keystone.Limit(25))`.
The standalone `test/requirements/aggregates` suite exercises this collection API
using the reusable `models.AggregateSample` source model.

`Count(alias)` counts entities; `CountProperty(property, alias)` counts non-null
values; `CountDistinct(property, alias)` counts distinct non-null values.
`Sum`, `Min`, `Max`, and `Avg` take `(property, alias)`.
`Aggregate(function, property, alias, true)` supports distinct aggregation with
any supported protobuf function. `Having` accepts property filters, including
`And`, `Or`, and null checks. Group keys and aliases use existing property names
and struct tags when decoded with `QueryRows[T]`; pointer structs also work.
The optional `QueryRows[T]` helper decodes raw response rows into your own named
result type when desired.
`IncludeTotal` counts groups in aggregate mode. `Explain()` requests execution
details. The server validates aggregate types and query combinations.

`QueryPages` and `QueryPagesWithRequest` support both entity and aggregate queries.
`Response()` holds only the latest page; `Next(ctx)` replaces it and returns
`io.EOF` at exhaustion. Failed fetches retain the page and cursor for retry.
Entity collections append entity pages through `Query`; use their `Aggregate`
method to create an aggregate collection.

`SaveResultSet()` saves ordered membership or aggregate rows for ten minutes on
the current server node, up to 20,000 results. It must start without a cursor.
The response exposes `ResultSet` and `ResultSetExpires`. Pagers and collections
switch automatically to token-based reads when advancing a saved query. Entity
properties are loaded afresh for each page; aggregate rows are saved values.
Keep requests on the same node. Tokens are bound to the original authorization,
expire, and are lost on restart; an unavailable token requires executing a new
query. Schema changes can also invalidate saved results.

To reopen a saved result set or manually fetch a page:

```go
response, err := actor.Query(ctx, "",
    keystone.ResultSet(token),
    keystone.Cursor(cursor), // omit to start from the beginning
    keystone.Limit(25),
    keystone.IncludeTotal(),
)
```

Only cursor, limit, total and explain controls may accompany a result token.
The entity type argument is ignored for saved reads. Filters, projections and
sorting belong to the original saved query and must not be sent again.
