package flickr

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	_ "log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SPRCallbackFunc func(StandardPhotoResponse) error

type FlickrAuthAPI struct {
	API
	Key      string
	Secret   string
	client   *http.Client
	throttle <-chan time.Time
}

func NewFlickrAuthAPI(key string, secret string) (API, error) {

	// https://github.com/golang/go/wiki/RateLimiting

	rate := time.Second / 10
	throttle := time.Tick(rate)

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	cl := &http.Client{Transport: tr}

	api := FlickrAuthAPI{
		Key:      key,
		Secret:   secret,
		throttle: throttle,
		client:   cl,
	}

	return &api, nil
}

func (api *FlickrAuthAPI) ExecuteMethod(method string, params url.Values) ([]byte, error) {

	params.Set("method", method)
	rsp, err := api.Call(params)

	if err != nil {
		return nil, err
	}

	if rsp.StatusCode != 200 {
		return nil, errors.New(rsp.Status)
	}

	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	// {"stat":"fail","code":96,"message":"Invalid signature"}
	// PLEASE FOR TO BE MAKING-AND-RETURNING A PACKAGE SPECIFIC
	// ERROR THINGY (20171204/thisisaaronland)

	stat := gjson.GetBytes(body, "stat")

	if !stat.Exists() {
		return nil, errors.New("Unable to determine response status")
	}

	if stat.String() != "ok" {

		errcode := gjson.GetBytes(body, "code")
		errmsg := gjson.GetBytes(body, "message")

		if !errcode.Exists() || !errmsg.Exists() {
			return nil, errors.New("Unable to parse error reponse")
		}

		msg := fmt.Sprintf("%d %s", errcode, errmsg)
		return nil, errors.New(msg)
	}

	return body, nil
}

func (api *FlickrAuthAPI) ExecuteMethodPaginated(method string, params url.Values, cb SPRCallbackFunc) error {

	page := 1
	pages := 0

	/*

		done_ch := make(chan bool)
		ticker := time.NewTicker(time.Second * 5)

		go func() {

			for range ticker.C {

				select {
				case <- done_ch:
					log.Printf("%s (%s) DONE\n", method, params.Get("woe_id"))
					return
				default:
					log.Printf("%s (%s) page %d/%d\n", method, params.Get("woe_id"), page, pages)
				}
			}

		}()

		defer func() {
			done_ch <- true
		}()

	*/

	for {

		params.Set("page", strconv.Itoa(page))

		rsp, err := api.ExecuteMethod(method, params)

		if err != nil {
			return err
		}

		var spr StandardPhotoResponse

		err = json.Unmarshal(rsp, &spr)

		if err != nil {
			return err
		}

		err = cb(spr)

		if err != nil {
			return err
		}

		pages = spr.Photos.Pages
		page += 1

		if pages == 0 || page > pages {
			break
		}
	}

	return nil
}

func (api FlickrAuthAPI) Call(params url.Values) (*http.Response, error) {

	params.Set("format", "json")
	params.Set("nojsoncallback", "1")
	params.Set("api_key", api.Key)

	// sig := api.Sign(params)
	// params.Set("api_sig", sig)

	url := "https://api.flickr.com/services/rest/"

	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return nil, err
	}

	// log.Printf("%s?%s\n", url, params.Encode())

	req.URL.RawQuery = params.Encode()

	<-api.throttle

	rsp, err := api.client.Do(req)

	// log.Println(req.URL, rsp.Status)
	return rsp, err
}

// copied from https://github.com/toomore/lazyflickrgo

func (api *FlickrAuthAPI) Sign(args url.Values) string {

	keySortedList := make([]string, len(args))
	var loop int64
	for key := range args {
		keySortedList[loop] = key
		loop++
	}
	sort.Strings(keySortedList)
	hashList := make([]string, len(args)*2)
	for i, val := range keySortedList {
		hashList[2*i] = val
		hashList[2*i+1] = args.Get(val)
	}

	hashstring := fmt.Sprintf("%s%s", api.Secret, strings.Join(hashList, ""))
	return fmt.Sprintf("%x", md5.Sum([]byte(hashstring)))
}
