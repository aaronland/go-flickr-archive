package archive

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thisisaaronland/go-flickr-archive/flickr"
	"github.com/tidwall/gjson"
	"log"
	"net/url"
	"strconv"
	"time"
)

type Archive struct {
	User User
	API  flickr.API
}

func NewArchiveForUser(api flickr.API, username string) (*Archive, error) {

	params := url.Values{}
	params.Set("username", username)

	rsp, err := api.ExecuteMethod("flickr.people.findByUsername", params)

	if err != nil {
		return nil, err
	}

	r := gjson.GetBytes(rsp, "user.nsid")

	if !r.Exists() {
		return nil, errors.New("can't find NSID")
	}

	nsid := r.String()

	params2 := url.Values{}
	params2.Set("user_id", nsid)

	rsp2, err := api.ExecuteMethod("flickr.people.getInfo", params2)

	if err != nil {
		return nil, err
	}

	r2 := gjson.GetBytes(rsp2, "person.photos.firstdate._content")

	if !r2.Exists() {
		return nil, errors.New("can't find NSID")
	}

	first := r2.Int()
	dt := time.Unix(first, 0)

	user := User{
		Username:   username,
		NSID:       nsid,
		FirstPhoto: dt,
	}

	arch := Archive{
		User: user,
		API:  api,
	}

	return &arch, nil
}

func (arch Archive) PhotosForDay(dt time.Time) error {

	// because time.Format() is just so weird...

	y, m, d := dt.Date()
	ymd := fmt.Sprintf("%04d-%02d-%02d", y, m, d)

	min_date := fmt.Sprintf("%s 00:00:00", ymd)
	max_date := fmt.Sprintf("%s 23:59:59", ymd)

	page := 1
	pages := 0

	for {

		params := url.Values{}
		params.Set("min_upload_date", min_date)
		params.Set("max_upload_date", max_date)
		params.Set("user_id", arch.User.NSID)
		params.Set("page", strconv.Itoa(page))

		rsp, err := arch.API.ExecuteMethod("flickr.people.getPhotos", params)

		if err != nil {
			return err
		}

		var spr flickr.StandardPhotoResponse

		err = json.Unmarshal(rsp, &spr)

		if err != nil {
			return err
		}

		for _, ph := range spr.Photos.Photos {
			arch.ArchivePhoto(ph)
		}

		pages = spr.Photos.Pages

		/*
			str_total := spr.Photos.Total
			total, err := strconv.Atoi(str_total)

			if err != nil {
				return err
			}

			log.Printf("page %d (of %d) %d\n", page, pages, total)
		*/

		if pages == 0 || pages == page {
			break
		}

		page += 1
	}

	return nil
}

func (arch Archive) ArchivePhoto(ph flickr.StandardPhotoResponsePhoto) error {

	// make ROOT/USER/pubic|private/YYYY/MM/DD/PHOTO_ID

	arch.ArchivePhotoInfo(ph)
	arch.ArchivePhotoSizes(ph)
	return nil
}

func (arch Archive) ArchivePhotoInfo(ph flickr.StandardPhotoResponsePhoto) error {

	// https://www.flickr.com/services/api/flickr.photos.getInfo.html

	params := url.Values{}
	params.Set("photo_id", ph.ID)
	params.Set("secret", ph.Secret)

	rsp, err := arch.API.ExecuteMethod("flickr.photos.getInfo", params)

	if err != nil {
		return err
	}

	log.Println(string(rsp))
	return nil
}

func (arch Archive) ArchivePhotoSizes(ph flickr.StandardPhotoResponsePhoto) error {

	// https://www.flickr.com/services/api/flickr.photos.getSizes.html

	params := url.Values{}
	params.Set("photo_id", ph.ID)

	rsp, err := arch.API.ExecuteMethod("flickr.photos.getSizes", params)

	if err != nil {
		return err
	}

	log.Println(string(rsp))
	return nil
}