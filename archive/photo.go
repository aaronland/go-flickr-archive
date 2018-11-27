package archive

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-storage"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
)

func archive_photos(api flickr.API, store storage.Store, photo_ids ...string) error {

	done_ch := make(chan bool)
	err_ch := make(chan error)

	for _, str_id := range photo_ids {

		go func(api flickr.API, store storage.Store, str_id string, done_ch chan bool, err_ch chan error) {

			defer func() {
				done_ch <- true
			}()

			err := archive_photo(api, store, str_id)

			if err != nil {
				err_ch <- err
			}

		}(api, store, str_id, done_ch, err_ch)
	}

	remaining := len(photo_ids)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			log.Println(e)
		default:
			// pass
		}
	}

	return nil
}

func archive_photo(api flickr.API, store storage.Store, str_id string) error {

	photo_id, err := strconv.ParseInt(str_id, 10, 64)

	if err != nil {
		return err
	}

	info_params := url.Values{}
	info_params.Set("photo_id", str_id)

	info, info_err := api.ExecuteMethod("flickr.photos.getInfo", info_params)

	if info_err != nil {
		return info_err
	}

	photo_rsp := gjson.GetBytes(info, "photo.id")

	if !photo_rsp.Exists() {
		return errors.New("Unable to determine photo ID")
	}

	if photo_rsp.Int() != photo_id {
		return errors.New("Mismatched photo ID")
	}

	secret_rsp := gjson.GetBytes(info, "photo.originalsecret")

	if !secret_rsp.Exists(){
		secret_rsp = gjson.GetBytes(info, "photo.secret")
	}

	if !secret_rsp.Exists(){
		return errors.New("Unable to determine photo secret")
	}

	secret := secret_rsp.String()
	
	sizes_params := url.Values{}
	sizes_params.Set("photo_id", str_id)

	sizes, sizes_err := api.ExecuteMethod("flickr.photos.getSizes", sizes_params)

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
	img_path := fmt.Sprintf("%d/%s", photo_id, img_fname)

	err = store.Put(img_path, img_rsp.Body)

	if err != nil {
		return err
	}
	
	info_path := fmt.Sprintf("%d/%d_%s_i.json", photo_id, photo_id, secret)
	
	info_r := bytes.NewReader(info)
	info_fh := ioutil.NopCloser(info_r)

	err = store.Put(info_path, info_fh)

	if err != nil {
		return err
	}

	return nil
}
