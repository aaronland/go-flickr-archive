package main

import (
	"flag"
	"github.com/aaronland/go-flickr-archive/archive"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/spr"
	"github.com/aaronland/go-storage"
	"github.com/whosonfirst/go-whosonfirst-cli/flags"
	"log"
	"net/url"
)

func main() {

	var key = flag.String("api-key", "", "...")
	var secret = flag.String("api-secret", "", "...")

	var storage_dsn = flag.String("storage", "", "...")

	var params flags.KeyValueArgs
	flag.Var(&params, "param", "...")

	flag.Parse()

	api, err := flickr.NewFlickrAuthAPI(*key, *secret)

	if err != nil {
		log.Fatal(err)
	}

	store, err := storage.NewFSStore(*storage_dsn)

	if err != nil {
		log.Fatal(err)
	}

	opts := archive.DefaultArchiveOptions()
	arch, err := archive.NewArchivist(api, store, opts)

	if err != nil {
		log.Fatal(err)
	}

	method := "flickr.photos.search"

	query := url.Values{}

	for _, p := range params {
		query.Set(p.Key, p.Value)
	}

	err = spr.ArchiveSPR(api, arch, method, query)

	if err != nil {
		log.Fatal(err)
	}
}
