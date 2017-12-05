package render

import (
	"bytes"
	"github.com/thisisaaronland/go-flickr-archive/assets"
	"github.com/thisisaaronland/go-flickr-archive/index"
	"github.com/thisisaaronland/go-flickr-archive/util"
	"html/template"
	"os"
	"path/filepath"
	"sync"
)

type StaticRender struct {
	Render
	Root string
}

func (r *StaticRender) RenderIndex(idx index.Index) error {

	for _, item := range idx.Items() {

		prev := idx.PreviousItem(item)
		next := idx.NextItem(item)

		err := r.RenderItem(item, prev, next)

		if err != nil {
			return err
		}
	}

	return nil
}

func (archive *StaticArchive) RenderItem(item index.IndexItem, prev index.IndexItem, next index.IndexItem) error {

	tpl, err := assets.Asset("templates/html/photo.html")

	if err != nil {
		return err
	}

	t, err := template.New("photo").Parse(string(tpl))

	if err != nil {
		return err
	}

	data := struct {
		Photo    index.IndexItem
		Next     index.IndexItem
		Previous index.IndexItem
	}{
		Photo:    item,
		Next:     next,
		Previous: prev,
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, data)

	if err != nil {
		return err
	}

	index_path := item.IndexPath()
	return util.WriteFile(index_path, buf.Bytes())
}
