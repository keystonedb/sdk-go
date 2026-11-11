# Entity alias requirements

Registered in the normal requirements runner as **Entity Aliases** and
**Entity Alias Destruction**. The latter requires the server's normal entity backup
and object-storage configuration and destroys only a customer created by the test.

The scenarios assert:

- Type-based SDK calls infer the owner from the actor; cross-app calls use
  `WithAliasOwner(vendor, app)` rather than constructing protobuf schema keys.
- Idempotent attachment, resolution, listing, and loading customer properties plus
  aliases by canonical ID or external alias.
- Removal, repeated removal, reactivation, conflicting ownership, and removal of
  an absent alias without reserving it.
- Routing-key, linked vendor/app, workspace and owning-schema isolation.
- Concurrent claims: exactly one winner, and only that customer's listing changes.
- Hashed-ID entities use canonical IDs in alias operations and load by alias.
- A linked app attaches, removes, reattaches and resolves its own alias without
  schema opt-in, then stores the canonical ID on its own record.
- Linked apps cannot attach/remove/resolve another app's aliases, list all aliases,
  or retrieve the owning customer by alias. Schemas need no opt-in.
- Missing identities/entities, malformed selectors, missing fields and size limits.
- Destruction removes resolution/list visibility but keeps ownership reserved.

To run only these scenarios against a **disposable, migrated test server**:

```sh
KEYSTONE_ALIAS_TEST_ENDPOINT=127.0.0.1:50051 \
KEYSTONE_ALIAS_TEST_TOKEN=test-access-token \
KEYSTONE_ALIAS_TEST_WORKSPACE=tt \
KEYSTONE_ALIAS_TEST_DESTROY=1 \
go test ./test/requirements/aliases -run TestAliasesAgainstServer -v -count=1
```

The test connection uses plaintext gRPC, suitable for a local/disposable endpoint.
The owner is `testing/tester`; the linked actor is `ven2/ap2`, and an unauthorized
actor is `ven2/other`. The environment must accept the supplied token for these apps,
as the existing requirements setup does with its development auth bypass. Workspace
`zz` is used to assert isolation. Never use production credentials or data.

Without `KEYSTONE_ALIAS_TEST_ENDPOINT`, `go test ./...` skips the server integration
case; it still runs SDK unit tests. Without `KEYSTONE_ALIAS_TEST_DESTROY=1`, the targeted
integration test runs the eight core scenarios only. The normal `go run ./test` runner
always includes the destruction requirement.

Tests remove their active aliases, but intentionally leave permanent ownership claims
and created test entities. Run against disposable storage. Apply Keystone migration
`012-entity-aliases.cql` before running against Cassandra/Scylla.
