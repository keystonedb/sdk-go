package keystone

// FindOption is an interface for options to be applied to a find request
type FindOption interface {
	Apply(config *filterRequest)
}
