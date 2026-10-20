package keystone

type withLimit struct {
	perPage    int32
	pageNumber int32
}

func (f withLimit) Apply(config *filterRequest) {
	config.PerPage = f.perPage
	config.PageNumber = f.pageNumber
}

// Limit sets the result size and an optional page number for Find-style APIs.
// Query supports only page zero; use Cursor to fetch subsequent pages.
func Limit(perPage int32, pageNumber ...int32) FindOption {
	var page int32
	if len(pageNumber) > 0 {
		page = pageNumber[0]
	}
	return withLimit{perPage: perPage, pageNumber: page}
}
