package photo

import (
	"strconv"
)

type Photo interface {
	Id() int64
}

type Archivist interface {
	ArchivePhotos(...Photo) error
}

type FlickrPhoto struct {
	Photo `json:",omitempty"`
	ID    int64 `json:"id"`
}

func NewFlickrPhotoFromString(str_id string) (Photo, error) {

	id, err := strconv.ParseInt(str_id, 10, 64)

	if err != nil {
		return nil, err
	}

	return NewFlickrPhoto(id)
}

func NewFlickrPhoto(id int64) (Photo, error) {

	ph := FlickrPhoto{
		ID: id,
	}

	return &ph, nil
}

func (ph *FlickrPhoto) Id() int64 {
	return ph.ID
}
