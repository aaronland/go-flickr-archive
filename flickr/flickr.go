package flickr

import (
       "net/http"
       "net/url"
)

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

type WIPStandardPhotoResponse interface {
     Page()	      int
     Pages()	      int
     PerPage()	      int
     Total()	      int
     Photos()	      []WIPStandardPhoto
}

type WIPStandardPhoto interface {
     ID()	      int64
     Owner()	      string
     Secret()	      string
     Server()	      int
     Farm()	      int
     Title()	      string
     IsPublic()	      bool
     IsPrivate()      bool
     IsFamily()	      bool
     PhotoPage()      url.URL
     PhotoURL()      url.URL     
}

type API interface {
	ExecuteMethod(string, url.Values) ([]byte, error)
	Call(url.Values) (*http.Response, error)
}
