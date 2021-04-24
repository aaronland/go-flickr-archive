package common

import (
	_ "context"
	"fmt"
	"github.com/aaronland/go-flickr-api/client"	
	"github.com/aaronland/go-flickr-archive"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
	"github.com/aaronland/go-flickr-archive/user"
	"net/url"
	"time"
)

func ArchivePhotosForUser(arch archive.Archivist, cl client.Client, u user.User) error {

	query := url.Values{}
	query.Set("user_id", u.ID())

	dt := u.DateFirstPhoto()

	for {

		err := ArchivePhotosWithSearchForDay(arch, cl, query, dt)

		if err != nil {
			return err
		}

		dt = dt.AddDate(0, 0, 1)
		today := time.Now()

		if dt.After(today) {
			break
		}
	}

	return nil
}

func ArchivePhotosWithSearchForDay(arch archive.Archivist, cl client.Client, query url.Values, dt time.Time) error {

	// because time.Format() is just so weird...

	y, m, d := dt.Date()
	ymd := fmt.Sprintf("%04d-%02d-%02d", y, m, d)

	min_date := fmt.Sprintf("%s 00:00:00", ymd)
	max_date := fmt.Sprintf("%s 23:59:59", ymd)

	query.Set("min_upload_date", min_date)
	query.Set("max_upload_date", max_date)

	return ArchivePhotosWithSearch(arch, cl, query)
}

func ArchivePhotosWithSearch(arch archive.Archivist, cl client.Client, query url.Values) error {

	method := "flickr.photos.search"
	return ArchivePhotosWithSPR(arch, cl, method, query)
}

func ArchivePhotosWithSPR(arch archive.Archivist, cl client.Client, method string, query url.Values) error {

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	cb := func(spr flickr.StandardPhotoResponse) error {

		photos := make([]photo.Photo, 0)

		for _, spr_ph := range spr.Photos.Photos {

			ph, err := photo.NewFlickrPhotoFromString(spr_ph.ID)

			if err != nil {
				return err
			}

			photos = append(photos, ph)
		}

		return arch.ArchivePhotos(cl, photos...)
	}

	return cl.ExecuteMethodPaginated(method, query, cb)
}
