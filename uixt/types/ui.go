package types

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

func (s Size) IsNil() bool {
	return s.Width == 0 && s.Height == 0
}
