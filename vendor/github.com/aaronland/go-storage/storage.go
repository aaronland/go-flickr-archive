package storage

// I am not convinced by this - it feels like a generic POSIX/UNIX
// style filesystem interface might be better... (20181019/thisisaaronland)

// https://github.com/spf13/afero - no S3 backend
// https://blog.gopheracademy.com/advent-2015/afero-a-universal-filesystem-library/
// https://github.com/usmanhalalit/gost
// https://github.com/src-d/go-billy - no S3 backend
// https://github.com/graymeta/stow - bespoke container/item interface
// https://github.com/src-d/go-billy - requires FUE

import (
	"io"
)

type WalkFunc func(string, ...interface{}) error

type Store interface {
	Get(string) (io.ReadCloser, error)
	Put(string, io.ReadCloser) error
	Delete(string) error
	Exists(string) (bool, error)
	Walk(WalkFunc) error
	URI(string) string
	Open(string) (io.WriteCloser, error)
}
