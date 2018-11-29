package archive

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
	"github.com/aaronland/go-flickr-archive/spr"
	"github.com/aaronland/go-flickr-archive/user"
	"github.com/aaronland/go-storage"
	"github.com/tidwall/gjson"
	"io/ioutil"
	_ "log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

type Archive interface {
	ArchivePhotos(...photo.Photo) (bool, []error)
	ArchivePhotosWithSearch(url.Values) (bool, []error)
}

type ArchiveOptions struct {
	ArchiveInfo     bool
	ArchiveSizes    bool
	ArchiveEXIF     bool
	ArchiveComments bool
	ArchiveRequest  bool
}

func DefaultArchiveOptions() *ArchiveOptions {

	opts := ArchiveOptions{
		ArchiveInfo:     true,
		ArchiveSizes:    false,
		ArchiveEXIF:     false,
		ArchiveComments: false,
		ArchiveRequest:  false,
	}

	return &opts
}

type Archivist struct {
	Archive
	api     flickr.API
	store   storage.Store
	options *ArchiveOptions
}

func NewArchivist(api flickr.API, store storage.Store, opts *ArchiveOptions) (Archive, error) {

	arch := Archivist{
		api:     api,
		store:   store,
		options: opts,
	}

	return &arch, nil
}

func ArchivePhotosForUser(arch Archive, u user.User) (bool, []error) {

	query := url.Values{}
	query.Set("user_id", u.ID())

	dt := u.DateFirstPhoto()

	for {

		ArchivePhotosWithSearchForDay(arch, query, dt)

		dt = dt.AddDate(0, 0, 1)
		today := time.Now()

		if dt.After(today) {
			break
		}
	}

	return true, nil
}

func ArchivePhotosWithSearchForDay(arch Archive, query url.Values, dt time.Time) (bool, []error) {

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// because time.Format() is just so weird...

	y, m, d := dt.Date()
	ymd := fmt.Sprintf("%04d-%02d-%02d", y, m, d)

	min_date := fmt.Sprintf("%s 00:00:00", ymd)
	max_date := fmt.Sprintf("%s 23:59:59", ymd)

	query.Set("min_upload_date", min_date)
	query.Set("max_upload_date", max_date)

	return arch.ArchivePhotosWithSearch(query)
}

// see this? it's an instance method mostly because I haven't decided whether or
// not to make 'API()' part of the interface... (20181129/thisisaaronland)

func (arch *Archivist) ArchivePhotosWithSearch(query url.Values) (bool, []error) {

	method := "flickr.photos.search"
	err := spr.ArchiveSPR(arch.api, arch, method, query)

	if err != nil {
	   return false, []error{ err }
	}

	return true, nil
}

func (arch *Archivist) ArchivePhotos(photos ...photo.Photo) (bool, []error) {

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, ph := range photos {

		go func(ph photo.Photo, done_ch chan bool, err_ch chan error) {

			defer func() {
				done_ch <- true
			}()

			err := arch.archivePhoto(ph)

			if err != nil {
				err_ch <- err
			}

		}(ph, done_ch, err_ch)
	}

	remaining := len(photos)
	arch_errors := make([]error, 0)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			arch_errors = append(arch_errors, e)
		default:
			// pass
		}
	}

	if len(arch_errors) > 0 {
		return false, arch_errors
	}

	return true, nil
}

func (arch *Archivist) archivePhoto(ph photo.Photo) error {

	str_id := strconv.FormatInt(ph.Id(), 10)

	info_params := url.Values{}
	info_params.Set("photo_id", str_id)

	info, info_err := arch.api.ExecuteMethod("flickr.photos.getInfo", info_params)

	if info_err != nil {
		return info_err
	}

	secret_rsp := gjson.GetBytes(info, "photo.originalsecret")

	if !secret_rsp.Exists() {
		secret_rsp = gjson.GetBytes(info, "photo.secret")
	}

	if !secret_rsp.Exists() {
		return errors.New("Unable to determine photo secret")
	}

	secret := secret_rsp.String()

	sizes_params := url.Values{}
	sizes_params.Set("photo_id", str_id)

	sizes, sizes_err := arch.api.ExecuteMethod("flickr.photos.getSizes", sizes_params)

	if sizes_err != nil {
		return sizes_err
	}

	photo_url := ""

	possible_sizes := []string{
		"Original",
		"Large 2048",
		"Large 1600",
		"Large",
		"Medium 800",
		"Medium 640",
		"Medium",
	}

	for _, label := range possible_sizes {

		path := fmt.Sprintf(`sizes.size.#[label="%s"]`, label)
		// log.Println("TRY", path)

		rsp := gjson.GetBytes(sizes, path)

		if !rsp.Exists() {
			continue
		}

		src := rsp.Get("source")
		photo_url = src.String()
		break
	}

	if photo_url == "" {
		return errors.New("Unable to determine photo URL")
	}

	img_rsp, err := http.Get(photo_url)

	if err != nil {
		return err
	}

	defer img_rsp.Body.Close()

	img_fname := filepath.Base(photo_url)
	img_path := fmt.Sprintf("%s/%s", str_id, img_fname)

	err = arch.store.Put(img_path, img_rsp.Body)

	if err != nil {
		return err
	}

	if arch.options.ArchiveInfo {

		info_path := fmt.Sprintf("%s/%s_%s_i.json", str_id, str_id, secret)

		info_r := bytes.NewReader(info)
		info_fh := ioutil.NopCloser(info_r)

		err = arch.store.Put(info_path, info_fh)

		if err != nil {
			return err
		}
	}

	if arch.options.ArchiveRequest {

		enc_ph, err := json.Marshal(ph)

		if err != nil {
			return nil
		}

		ph_r := bytes.NewReader(enc_ph)
		ph_fh := ioutil.NopCloser(ph_r)

		// should this have a secret? (20181127/thisisaaronland)

		ph_path := fmt.Sprintf("%s/%s_r.json", str_id, str_id)

		err = arch.store.Put(ph_path, ph_fh)

		if err != nil {
			return err
		}
	}

	return nil
}
