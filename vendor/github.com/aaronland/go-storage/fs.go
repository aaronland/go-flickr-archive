package storage

import (
	"errors"
	"github.com/aaronland/go-string/dsn"
	"github.com/whosonfirst/walk"
	"io"
	_ "log"
	"os"
	"path/filepath"
	_ "strconv"
)

type FSFile struct {
	io.WriteCloser
	fh *os.File
}

func (f *FSFile) Write(b []byte) (int, error) {
	return f.fh.Write(b)
}

func (f *FSFile) WriteString(b string) (int, error) {
	return f.fh.Write([]byte(b))
}

func (f *FSFile) Close() error {
	return f.fh.Close()
}

type FSStore struct {
	Store
	root       string
	file_perms os.FileMode
	dir_perms  os.FileMode
}

func NewFSStore(str_dsn string) (Store, error) {

	dsn_map, err := dsn.StringToDSNWithKeys(str_dsn, "root")

	if err != nil {
		return nil, err
	}

	abs_root, err := filepath.Abs(dsn_map["root"])

	if err != nil {
		return nil, err
	}

	/*
	file_perms := 0644
	dir_perms := 0755

	str_fileperms, ok := dsn_map["file_perms"]

	if ok {

		perms, err := strconv.ParseUint(str_fileperms, 10, 32)

		if err != nil {
			return nil, errors.New("Invalid file permissions")
		}

		file_perms = perms
	}

	str_dirperms, ok := dsn_map["dir_perms"]

	if ok {

		perms, err := strconv.ParseUint(str_dirperms, 10, 32)

		if err != nil {
			return nil, errors.New("Invalid directory permissions")
		}

		dir_perms = perms
	}
	*/
	
	s := FSStore{
		root:       abs_root,
		file_perms: 0644,
		dir_perms:  0755,
	}

	return &s, nil
}

func (s *FSStore) URI(k string) string {
	return filepath.Join(s.root, k)
}

func (s *FSStore) Get(k string) (io.ReadCloser, error) {

	path := filepath.Join(s.root, k)
	return os.Open(path)
}

func (s *FSStore) Open(k string) (io.WriteCloser, error) {

	path := filepath.Join(s.root, k)

	abs_path, err := filepath.Abs(path)

	if err != nil {
		return nil, err
	}

	root := filepath.Dir(abs_path)

	_, err = os.Stat(root)

	if os.IsNotExist(err) {

		err = os.MkdirAll(root, s.dir_perms)

		if err != nil {
			return nil, err
		}
	}

	return os.OpenFile(abs_path, os.O_RDWR|os.O_CREATE, s.file_perms)
}

func (s *FSStore) Put(k string, in io.ReadCloser) error {

	out, err := s.Open(k)

	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func (s *FSStore) Delete(k string) error {

	path := filepath.Join(s.root, k)

	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("Deleting directories not supported yet")
	}

	return os.Remove(path)
}

func (s *FSStore) Exists(k string) (bool, error) {

	path := filepath.Join(s.root, k)

	_, err := os.Stat(path)

	if os.IsNotExist(err) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *FSStore) Walk(user_cb WalkFunc) error {

	walk_cb := func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		return user_cb(path, info)
	}

	return walk.Walk(s.root, walk_cb)
}
