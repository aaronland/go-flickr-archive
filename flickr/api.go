package flickr

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
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
	Key    string
	Secret string
}

func NewFlickrAuthAPI(key string, secret string) (API, error) {

	api := FlickrAuthAPI{
		Key:    key,
		Secret: secret,
	}

	return &api, nil
}

func (api *FlickrAuthAPI) ExecuteMethod(method string, params url.Values) ([]byte, error) {

	params.Set("method", method)
	rsp, err := api.Call(params)

	if err != nil {
		return nil, err
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

		if pages == 0 || pages == page {
			break
		}

		page += 1
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

	tr := &http.Transport{
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	cl := &http.Client{Transport: tr}

	req, err := http.NewRequest("POST", url, nil)

	if err != nil {
		return nil, err
	}

	// log.Printf("%s?%s\n", url, params.Encode())

	req.URL.RawQuery = params.Encode()
	return cl.Do(req)
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
