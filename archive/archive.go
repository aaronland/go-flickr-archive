package archive

import (
	"context"
	"github.com/aaronland/go-flickr-archive/flickr"
	"time"
)

type Archive interface {
	ArchivePhotosForDay(time.Time) error
	ArchivePhoto(context.Context, flickr.StandardPhotoResponsePhoto) error
}
