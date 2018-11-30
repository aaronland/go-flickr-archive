package archive

import (
	"context"
	"github.com/aaronland/go-flickr-archive/photo"
)

type Archive interface {
	ArchivePhotos(context.Context, ...photo.Photo) error
	ArchivePhoto(context.Context, photo.Photo) error
}

type ArchiveOptions struct {
	ArchiveInfo     bool
	ArchiveSizes    bool
	ArchiveEXIF     bool
	ArchiveComments bool
	ArchiveRequest  bool
}

func DefaultArchiveOptions() *ArchiveOptions {

	opts := ArchiveOptions{
		ArchiveInfo:     true,
		ArchiveSizes:    false,
		ArchiveEXIF:     false,
		ArchiveComments: false,
		ArchiveRequest:  false,
	}

	return &opts
}
