package proto

// WithRouting sets the alias routing namespace and returns the same alias for
// chaining. An empty key selects the default namespace; it is not a wildcard.
func (a *EntityAlias) WithRouting(routingKey string) *EntityAlias {
	a.RoutingKey = routingKey
	return a
}
