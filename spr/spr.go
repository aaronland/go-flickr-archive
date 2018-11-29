package spr

// this is ppppprrrrrobably deprecated since (arch *Archivist) ArchivePhotosForSPR has been
// added but not sure yet... (20181129/thisisaaronland)

import (
	"errors"
	"github.com/aaronland/go-flickr-archive/archive"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
	"log"
	"net/url"
)

func ArchiveSPR(api flickr.API, arch archive.Archive, method string, query url.Values) error {

	cb := func(spr flickr.StandardPhotoResponse) error {

		photos := make([]photo.Photo, 0)

		for _, spr_ph := range spr.Photos.Photos {

			ph, err := photo.NewFlickrPhotoFromString(spr_ph.ID)

			if err != nil {
				return err
			}

			photos = append(photos, ph)
		}

		ok, errs := arch.ArchivePhotos(photos...)

		if !ok {

			for _, e := range errs {
				log.Println(e)
			}

			return errors.New("One or more photos failed to be archived")
		}

		return nil
	}

	return api.ExecuteMethodPaginated(method, query, cb)
}
