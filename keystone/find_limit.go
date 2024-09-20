package keystone

type withLimit struct {
	perPage    int32
	pageNumber int32
}

func (f withLimit) Apply(config *filterRequest) {
	config.PerPage = f.perPage
	config.PageNumber = f.pageNumber
}

func Limit(perPage, pageNumber int32) FindOption {
	return withLimit{perPage: perPage, pageNumber: pageNumber}
}
