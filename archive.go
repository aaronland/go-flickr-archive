package archive

import (
	"context"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
)

type Archivist interface {
	ArchivePhotos(flickr.API, ...photo.Photo) error
	ArchivePhoto(context.Context, flickr.API, photo.Photo) error
}
