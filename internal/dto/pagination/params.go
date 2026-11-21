package pagination

type Params struct {
	Cursor *Cursor
	Limit  int
}

func (params Params) WithLookahead() Params {
	params.Limit++
	return params
}
