package spr

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

/*

for user, by day stuff

func (archive *StaticArchive) ArchivePhotosForDay(dt time.Time) error {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// because time.Format() is just so weird...

	y, m, d := dt.Date()
	ymd := fmt.Sprintf("%04d-%02d-%02d", y, m, d)

	min_date := fmt.Sprintf("%s 00:00:00", ymd)
	max_date := fmt.Sprintf("%s 23:59:59", ymd)

	user_id := archive.User.ID()

*/
