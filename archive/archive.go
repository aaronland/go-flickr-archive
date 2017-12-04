package archive

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/facebookgo/atomicfile"
	"github.com/thisisaaronland/go-flickr-archive/flickr"
	"github.com/tidwall/gjson"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Archive struct {
	User User
	API  flickr.API
	Root string
}

func NewArchiveForUser(api flickr.API, username string, root string) (*Archive, error) {

	info, err := os.Stat(root)

	if os.IsNotExist(err) {
		return nil, err
	}

	if !info.IsDir() {
		return nil, errors.New("Archive root is not a directory")
	}

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
		Root: "",
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

	// https://www.flickr.com/services/api/flickr.photos.getInfo.html

	info_params := url.Values{}
	info_params.Set("photo_id", ph.ID)
	info_params.Set("secret", ph.Secret)

	info, info_err := arch.API.ExecuteMethod("flickr.photos.getInfo", info_params)

	if info_err != nil {
		return info_err
	}

	// { "photo": { "id": "38117183734", "secret": "2e8c5fce11", "server": "4520", "farm": 5, "dateuploaded": "1512404114", "isfavorite": 0, "license": 0, "safety_level": 0, "rotation": 0,
	//     "visibility": { "ispublic": 1, "isfriend": 0, "isfamily": 0 },
	//     "dates": { "posted": "1512404114", "taken": "2017-12-03 10:57:38", "takengranularity": 0, "takenunknown": 1, "lastupdate": "1512404117" }, "views": 0,

	photo_id := gjson.GetBytes(info, "photo.id")

	if !photo_id.Exists() {
		return errors.New("Unable to determin photo ID")
	}

	date_taken := gjson.GetBytes(info, "photo.dates.taken")

	if !date_taken.Exists() {
		return errors.New("Unable to determine date taken")
	}

	ymd_taken := strings.Split(date_taken.String(), " ")
	ymd := ymd_taken[0]

	dt, err := time.Parse("2006-01-02", ymd)

	if err != nil {
		return err
	}

	is_public := gjson.GetBytes(info, "photo.visibility.ispublic")

	if !is_public.Exists() {
		return errors.New("Unable to determine visibility")
	}

	visibility := "private" // default

	if is_public.Bool() {
		visibility = "public"
	}

	var secret gjson.Result

	if visibility == "public" {
		secret = gjson.GetBytes(info, "photo.secret")
	} else {
		secret = gjson.GetBytes(info, "photo.originalsecret") // is that right?
	}

	if !secret.Exists() {
		return errors.New("Unable to determine secret")
	}

	root := filepath.Join(arch.Root, arch.User.Username)
	root = filepath.Join(root, visibility)
	root = filepath.Join(root, fmt.Sprintf("%04d", dt.Year()))
	root = filepath.Join(root, fmt.Sprintf("%02d", dt.Month()))
	root = filepath.Join(root, fmt.Sprintf("%02d", dt.Day()))
	root = filepath.Join(root, fmt.Sprintf("%d", photo_id.Int()))

	info_fname := fmt.Sprintf("%d_%s_i.json", photo_id.Int(), secret.String())
	info_path := filepath.Join(root, info_fname)

	log.Println("ARCHIVE", info_path)

	_, err = os.Stat(root)

	if os.IsNotExist(err) {

		err = os.MkdirAll(root, 0755) // configurable perms

		if err != nil {
			return err
		}
	}

	err = arch.WriteFile(info_path, info)

	if err != nil {
		return err
	}

	sz_params := url.Values{}
	sz_params.Set("photo_id", ph.ID)

	_, sz_err := arch.API.ExecuteMethod("flickr.photos.getSizes", sz_params)

	if sz_err != nil {
		return sz_err
	}

	// { "sizes": { "canblog": 0, "canprint": 0, "candownload": 1,
	// "size": [
	// { "label": "Square", "width": 75, "height": 75, "source": "https:\/\/farm5.staticflickr.com\/4515\/24960711428_7d25eac274_s.jpg", "url": "https:\/\/www.flickr.com\/photos\/138795394@N02\/24960711428\/sizes\/sq\/", "media": "photo" },

	// make ROOT/USER/pubic|private/YYYY/MM/DD/PHOTO_ID
	// write INFO to disk as PHOTO_ID_ORIGINALSECRET_i.json
	// fetch all the sizes and write to disk

	return nil
}

func (arch Archive) WriteFile(path string, body []byte) error {

	fh, err := atomicfile.New(path, 0644) // custom perms?

	if err != nil {
		return err
	}

	_, err = fh.Write(body)

	if err != nil {
		fh.Abort()
		return err
	}

	err = fh.Close()

	if err != nil {
		fh.Abort()
		return err
	}

	return nil
}
