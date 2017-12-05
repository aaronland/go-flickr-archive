package archive

import (
	"context"
	"github.com/thisisaaronland/go-flickr-archive/flickr"
	"time"
)

type Archive interface {
	ArchivePhotosForDay(time.Time) error
	ArchivePhoto(context.Context, flickr.StandardPhotoResponsePhoto) error
}

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
}
