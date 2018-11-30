package archive

import (
	_ "context"
	"fmt"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
	"github.com/aaronland/go-flickr-archive/user"
	"net/url"
	"time"
)

func ArchivePhotosForUser(arch Archive, api flickr.API, u user.User) error {

	query := url.Values{}
	query.Set("user_id", u.ID())

	dt := u.DateFirstPhoto()

	for {

		err := ArchivePhotosWithSearchForDay(arch, api, query, dt)

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

func ArchivePhotosWithSearchForDay(arch Archive, api flickr.API, query url.Values, dt time.Time) error {

	// because time.Format() is just so weird...

	y, m, d := dt.Date()
	ymd := fmt.Sprintf("%04d-%02d-%02d", y, m, d)

	min_date := fmt.Sprintf("%s 00:00:00", ymd)
	max_date := fmt.Sprintf("%s 23:59:59", ymd)

	query.Set("min_upload_date", min_date)
	query.Set("max_upload_date", max_date)

	return ArchivePhotosWithSearch(arch, api, query)
}

func ArchivePhotosWithSearch(arch Archive, api flickr.API, query url.Values) error {

	method := "flickr.photos.search"
	return ArchivePhotosWithSPR(arch, api, method, query)
}

func ArchivePhotosWithSPR(arch Archive, api flickr.API, method string, query url.Values) error {

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

		return arch.ArchivePhotos(api, photos...)
	}

	return api.ExecuteMethodPaginated(method, query, cb)
}
