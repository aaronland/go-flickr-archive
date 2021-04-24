package archive

import (
	"context"
	"github.com/aaronland/go-flickr-api/client"
	"github.com/aaronland/go-flickr-archive/photo"
)

type Archivist interface {
	ArchivePhotos(client.Client, ...photo.Photo) error
	ArchivePhoto(context.Context, client.Client, photo.Photo) error
}
