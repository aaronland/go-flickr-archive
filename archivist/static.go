package archivist

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aaronland/go-flickr-archive"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/aaronland/go-flickr-archive/photo"
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

type StaticArchivistOptions struct {
	ArchiveInfo       bool
	ArchiveSizes      bool
	ArchiveEXIF       bool
	ArchiveComments   bool
	ArchiveRequest    bool
	RequestsPerSecond int
	// Logger
	// Throttle
}

func DefaultStaticArchivistOptions() (*StaticArchivistOptions, error) {

	opts := StaticArchivistOptions{
		ArchiveInfo:       true,
		ArchiveSizes:      false,
		ArchiveEXIF:       false,
		ArchiveComments:   false,
		ArchiveRequest:    false,
		RequestsPerSecond: 10,
	}

	return &opts, nil
}

type StaticArchivist struct {
	archive.Archivist
	store    storage.Store
	options  *StaticArchivistOptions
	throttle <-chan time.Time
	client   *http.Client
}

func NewStaticArchivist(store storage.Store, opts *StaticArchivistOptions) (archive.Archivist, error) {

	// https://github.com/golang/go/wiki/RateLimiting

	rate := time.Second / time.Duration(opts.RequestsPerSecond)
	throttle := time.Tick(rate)

	// maybe make this an option?

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	client := &http.Client{Transport: tr}

	arch := StaticArchivist{
		store:    store,
		options:  opts,
		throttle: throttle,
		client:   client,
	}

	return &arch, nil
}

func (arch *StaticArchivist) ArchivePhotos(api flickr.API, photos ...photo.Photo) error {

	done_ch := make(chan bool)
	err_ch := make(chan error)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, ph := range photos {

		go func(ctx context.Context, ph photo.Photo, done_ch chan bool, err_ch chan error) {

			defer func() {
				done_ch <- true
			}()

			select {
			case <-ctx.Done():
				return
			default:
				// pass
			}

			err := arch.ArchivePhoto(ctx, api, ph)

			if err != nil {
				err_ch <- err
			}

		}(ctx, ph, done_ch, err_ch)
	}

	remaining := len(photos)

	for remaining > 0 {

		select {
		case <-done_ch:
			remaining -= 1
		case e := <-err_ch:
			return e
		default:
			// pass
		}
	}

	return nil
}

func (arch *StaticArchivist) ArchivePhoto(ctx context.Context, api flickr.API, ph photo.Photo) error {

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	<-arch.throttle

	str_id := strconv.FormatInt(ph.Id(), 10)

	info_params := url.Values{}
	info_params.Set("photo_id", str_id)

	info, info_err := api.ExecuteMethod("flickr.photos.getInfo", info_params)

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

	img_rsp, err := arch.client.Get(photo_url)

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
