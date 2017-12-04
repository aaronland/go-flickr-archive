package main

import (
	"flag"
	"github.com/thisisaaronland/go-flickr-archive/archive"
	"github.com/thisisaaronland/go-flickr-archive/flickr"
	"log"
	"time"
)

func main() {

	var key = flag.String("api-key", "", "...")
	var secret = flag.String("api-secret", "", "...")
	var username = flag.String("username", "", "...")

	flag.Parse()

	api, err := flickr.NewFlickrAuthAPI(*key, *secret)

	if err != nil {
		log.Fatal(err)
	}

	arch, err := archive.NewArchiveForUser(api, *username)

	if err != nil {
		log.Fatal(err)
	}

	dt := arch.User.FirstPhoto

	for {

		log.Println(dt.Format(time.RFC3339))
		arch.PhotosForDay(dt)

		dt = dt.AddDate(0, 0, 1)
		today := time.Now()

		if dt.After(today) {
			break
		}
	}
}