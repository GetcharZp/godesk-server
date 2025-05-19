package util

import godesk "github.com/getcharzp/godesk-serve/proto"

type Pager struct {
	Page int
	Size int
}

func NewPager(base *godesk.BaseRequest) *Pager {
	return &Pager{
		Page: int(base.Page),
		Size: int(base.Size),
	}
}

func (p *Pager) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Size
}

func (p *Pager) Limit() int {
	return p.Size
}
