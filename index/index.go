package index

import (
	"time"
)

type Index interface {
	Items() []IndexItem
	PreviousItem(IndexItem) IndexItem
	NextItem(IndexItem) IndexItem
}

type IndexItem interface {
	ID() int64
	Title() string
	Secret() string
	Date() time.Time
	RootPath() string
	ImagePath() string
	IndexPath() string
}
