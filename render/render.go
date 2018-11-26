package render

import (
	"time"
)

type RenderIndex interface {
	Items() []RenderIndexItem
	PreviousItem(RenderIndexItem) RenderIndexItem
	NextItem(RenderIndexItem) RenderIndexItem
}

type RenderIndexItem interface {
	ID() int64
	Title() string
	Secret() string
	Date() time.Time
	RootPath() string
	ImagePath() string
	IndexPath() string
}
