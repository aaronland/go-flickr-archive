package archive

import (
	_ "context"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
)

type Archive interface {
	ArchivePhotos(flickr.API, ...photo.Photo) error
	ArchivePhoto(flickr.API, photo.Photo) error
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
