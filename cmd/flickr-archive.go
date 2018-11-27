package main

import (
	"flag"
	"github.com/aaronland/go-flickr-archive/archive"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
	"github.com/aaronland/go-storage"
	"log"
	"os"
	"path/filepath"
	_ "time"
)

func main() {

	var key = flag.String("api-key", "", "...")
	var secret = flag.String("api-secret", "", "...")
	// var username = flag.String("username", "", "...")

	// please support other storage layers...x
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

	opts := archive.DefaultArchiveOptions()
	arch, err := archive.NewArchivist(api, store, opts)

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

	ok, errs := arch.ArchivePhotos(photos...)

	if !ok {

		for _, e := range errs {
			log.Println(e)
		}

		os.Exit(1)
	}

	// old code to be updated...

	/*


		u, err := user.NewArchiveUserForUsername(api, *username)

		arch, err := archive.NewStaticArchiveForUser(api, u, abs_root)

		if err != nil {
			log.Fatal(err)
		}

		dt := u.DateFirstPhoto()

		for {

			log.Println(dt.Format(time.RFC3339))
			arch.ArchivePhotosForDay(dt)

			dt = dt.AddDate(0, 0, 1)
			today := time.Now()

			if dt.After(today) {
				break
			}
		}
	*/
}
