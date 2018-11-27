package user

import (
	"errors"
	"github.com/aaronland/go-flickr-archive/flickr"
	"github.com/tidwall/gjson"
	"net/url"
	"time"
)

type User interface {
	Username() string
	ID() string
	DateFirstPhoto() time.Time
}

type ArchiveUser struct {
	User
	username string
	nsid     string
	first    time.Time // please rename me...
}

func NewArchiveUserForUsername(api flickr.API, username string) (User, error) {

	find_params := url.Values{}
	find_params.Set("username", username)

	found, err := api.ExecuteMethod("flickr.people.findByUsername", find_params)

	if err != nil {
		return nil, err
	}

	nsid := gjson.GetBytes(found, "user.nsid")

	if !nsid.Exists() {
		return nil, errors.New("can't find NSID")
	}

	info_params := url.Values{}
	info_params.Set("user_id", nsid.String())

	info, err := api.ExecuteMethod("flickr.people.getInfo", info_params)

	if err != nil {
		return nil, err
	}

	first := gjson.GetBytes(info, "person.photos.firstdate._content")

	if !first.Exists() {
		return nil, errors.New("can't find NSID")
	}

	first_ts := first.Int()
	dt := time.Unix(first_ts, 0)

	user := ArchiveUser{
		username: username,
		nsid:     nsid.String(),
		first:    dt,
	}

	return &user, nil
}

func (u *ArchiveUser) Username() string {
	return u.username
}

func (u *ArchiveUser) ID() string {
	return u.nsid
}

func (u *ArchiveUser) DateFirstPhoto() time.Time {
	return u.first
}
