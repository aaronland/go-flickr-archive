package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

type User struct {
	Username   string
	NSID       string
	FirstPhoto time.Time
}

// {"stat":"fail","code":96,"message":"Invalid signature"}

type StandardPhotoResponse struct {
	Photos StandardPhotoResponsePhotos `json:"photos"`
	Stat   string                      `json:"stat"`
}

type StandardPhotoResponsePhotos struct {
	Page    int                          `json:"page"`
	Pages   int                          `json:"pages"`
	PerPage int                          `json:"perpage"`
	Total   string                       `json:"total"` // see the way this is a string... what???
	Photos  []StandardPhotoResponsePhoto `json:"photo"` // see the way its 'photo' and not 'photos' ... yeah, that
}

type StandardPhotoResponsePhoto struct {
	ID       string `json:"id"` // string... what??
	Owner    string `json:"owner"`
	Secret   string `json:"secret"`
	Server   string `json:"server"` // string... what??
	Farm     int    `json:"farm"`
	Title    string `json:title"`
	IsPublic int    `json:ispublic"` // Y U NO bool
	IsFriend int    `json:isfriend"` // see above
	IsFamily int    `json:isfamily"` // see above
}

type API struct {
	Key    string
	Secret string
}

func (api API) CallAsBytes(params url.Values) ([]byte, error) {

	rsp, err := api.Call(params)

	if err != nil {
		return nil, err
	}

	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (api API) Call(params url.Values) (*http.Response, error) {

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

func (api API) Sign(args url.Values) string {

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

type Archive struct {
	User User
	API  API
}

func (arch Archive) PhotosForDay(dt time.Time) {

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
		params.Set("method", "flickr.people.getPhotos")
		params.Set("user_id", arch.User.NSID)
		params.Set("page", strconv.Itoa(page))

		rsp, err := arch.API.Call(params)

		if err != nil {
			log.Fatal(err)
		}

		defer rsp.Body.Close()

		body, err := ioutil.ReadAll(rsp.Body)

		if err != nil {
			log.Fatal(err)
		}

		var spr StandardPhotoResponse

		err = json.Unmarshal(body, &spr)

		if err != nil {
			log.Fatal(err)
		}

		for _, ph := range spr.Photos.Photos {
			arch.ArchivePhoto(ph)
		}

		str_total := spr.Photos.Total
		total, err := strconv.Atoi(str_total)

		if err != nil {
			log.Fatal(err)
		}

		pages = spr.Photos.Pages

		log.Printf("page %d (of %d) %d\n", page, pages, total)

		if pages == 0 || pages == page {
			break
		}

		page += 1
	}

}

func (arch Archive) ArchivePhoto(ph StandardPhotoResponsePhoto) error {

	// make ROOT/USER/pubic|private/YYYY/MM/DD/PHOTO_ID

	arch.ArchivePhotoInfo(ph)
	arch.ArchivePhotoSizes(ph)
	return nil
}

func (arch Archive) ArchivePhotoInfo(ph StandardPhotoResponsePhoto) error {

	// https://www.flickr.com/services/api/flickr.photos.getInfo.html

	params := url.Values{}
	params.Set("method", "flickr.photos.getInfo")
	params.Set("photo_id", ph.ID)
	params.Set("secret", ph.Secret)

	rsp, err := arch.API.Call(params)

	if err != nil {
		log.Fatal(err)
	}

	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)

	log.Println(string(body))
	return nil
}

func (arch Archive) ArchivePhotoSizes(ph StandardPhotoResponsePhoto) error {

	// https://www.flickr.com/services/api/flickr.photos.getSizes.html

	params := url.Values{}
	params.Set("method", "flickr.photos.getSizes")
	params.Set("photo_id", ph.ID)

	rsp, err := arch.API.Call(params)

	if err != nil {
		log.Fatal(err)
	}

	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)

	log.Println(string(body))
	return nil
}

func main() {

	var key = flag.String("api-key", "", "...")
	var secret = flag.String("api-secret", "", "...")
	var username = flag.String("username", "", "...")

	flag.Parse()

	api := API{
		Key:    *key,
		Secret: *secret,
	}

	params := url.Values{}
	params.Set("method", "flickr.people.findByUsername")
	params.Set("username", *username)

	rsp, err := api.CallAsBytes(params)

	if err != nil {
		log.Fatal(err)
	}

	r := gjson.GetBytes(rsp, "user.nsid")

	if !r.Exists() {
		log.Fatal("can't find NSID")
	}

	nsid := r.String()

	params2 := url.Values{}
	params2.Set("method", "flickr.people.getInfo")
	params2.Set("user_id", nsid)

	rsp2, err := api.CallAsBytes(params2)

	if err != nil {
		log.Fatal(err)
	}

	log.Println(string(rsp2))
	r2 := gjson.GetBytes(rsp2, "person.photos.firstdate._content")

	if !r2.Exists() {
		log.Fatal("can't find NSID")
	}

	first := r2.Int()
	dt := time.Unix(first, 0)

	user := User{
		Username:   *username,
		NSID:       nsid,
		FirstPhoto: dt,
	}

	arch := Archive{
		User: user,
		API:  api,
	}

	for {

		log.Println(dt.Format(time.RFC3339))
		arch.PhotosForDay(dt)

		dt = dt.AddDate(0, 0, 1)
		today := time.Now()

		if dt.After(today) {
			break
		}
	}
}
