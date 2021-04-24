package main

import (
	"flag"
	"github.com/aaronland/go-flickr-archive/archivist"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
	"github.com/aaronland/go-storage"
	"log"
	"path/filepath"
)

func main() {

	var key = flag.String("api-key", "", "...")
	var secret = flag.String("api-secret", "", "...")

	// please support other storage layers...
	var root = flag.String("root", "", "...")

	flag.Parse()

	abs_root, err := filepath.Abs(*root)

	if err != nil {
		log.Fatal(err)
	}

	api, err := flickr.NewFlickrAuthAPI(*key, *secret)

	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.NewFSStore(abs_root)

	if err != nil {
		log.Fatal(err)
	}

	opts, err := archivist.DefaultStaticArchivistOptions()

	if err != nil {
		log.Fatal(err)
	}

	arch, err := archivist.NewStaticArchivist(store, opts)

	if err != nil {
		log.Fatal(err)
	}

	photos := make([]photo.Photo, 0)

	for _, str_id := range flag.Args() {

		ph, err := photo.NewFlickrPhotoFromString(str_id)

		if err != nil {
			log.Fatal(err)
		}

		photos = append(photos, ph)
	}

	err = arch.ArchivePhotos(api, photos...)

	if err != nil {
		log.Fatal(err)
	}

}
