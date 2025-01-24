package keystone

type Error struct {
	ErrorCode    int32
	ErrorMessage string
	Suggestions  []string
	Extended     []string
}

func (e *Error) Error() string {
	return e.ErrorMessage
}
