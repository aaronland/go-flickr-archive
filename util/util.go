package util

import (
	"github.com/facebookgo/atomicfile"
	"io/ioutil"
	"net/http"
	"os"
)

func GetStore(remote string, local string) error {

	rsp, err := http.Get(remote)

	if err != nil {
		return err
	}

	defer rsp.Body.Close()

	body, err := ioutil.ReadAll(rsp.Body)

	if err != nil {
		return err
	}

	err = WriteFile(local, body)

	if err != nil {
		return err
	}

	return nil
}

func ReadFile(path string) ([]byte, error) {

	fh, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(fh)
}

func WriteFile(path string, body []byte) error {

	fh, err := atomicfile.New(path, 0644) // custom perms?

	if err != nil {
		return err
	}

	_, err = fh.Write(body)

	if err != nil {
		fh.Abort()
		return err
	}

	err = fh.Close()

	if err != nil {
		fh.Abort()
		return err
	}

	return nil
}
