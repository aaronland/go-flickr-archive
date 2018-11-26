package index

import (
	"fmt"
	"path/filepath"
	"time"
)

type StaticIndex struct {
	Index
	items       []IndexItem
	breadcrumbs map[int64][]int
}

func (i *StaticIndex) Items() []IndexItem {
	return i.items
}

func (i *StaticIndex) PreviousItem(item IndexItem) IndexItem {

	id := item.ID()

	nav, ok := i.breadcrumbs[id]

	if !ok {
		return nil
	}

	prev := nav[0]

	if prev == -1 {
		return nil
	}

	return i.items[prev]
}

func (i *StaticIndex) NextItem(item IndexItem) IndexItem {

	id := item.ID()

	nav, ok := i.breadcrumbs[id]

	if !ok {
		return nil
	}

	next := nav[1]

	if next == -1 {
		return nil
	}

	return i.items[next]
}

type StaticIndexItem struct {
	id     int64
	title  string
	secret string
	date   time.Time
}

func (i *StaticIndexItem) ID() int64 {
	return i.id
}

func (i *StaticIndexItem) Title() string {
	return i.title
}

func (i *StaticIndexItem) Secret() string {
	return i.secret
}

func (i *StaticIndexItem) Date() time.Time {
	return i.date
}

func (i *StaticIndexItem) RootPath() string {

	dt := i.Date()
	id := i.ID()

	root := ""

	root = filepath.Join(root, fmt.Sprintf("%04d", dt.Year()))
	root = filepath.Join(root, fmt.Sprintf("%02d", dt.Month()))
	root = filepath.Join(root, fmt.Sprintf("%02d", dt.Day()))
	root = filepath.Join(root, fmt.Sprintf("%d", id))

	return root
}

func (i *StaticIndexItem) ImagePath() string {

	root := i.RootPath()

	id := i.ID()
	secret := i.Secret()

	fname := fmt.Sprintf("%d_%s_z.jpg", id, secret)
	return filepath.Join(root, fname)
}

func (i *StaticIndexItem) IndexPath() string {

	root := i.RootPath()
	return filepath.Join(root, "index.html")
}

/*
func (r *StaticRender) IndexStaticArchive(path string) (Index, error) {

	items := make([]IndexItem, 0)
	mu := new(sync.Mutex)

	cb := func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		abs_path, err := filepath.Abs(path)

		if err != nil {
			// log.Println("PATH", abs_path, err)
			return nil
		}

		if !strings.HasSuffix(abs_path, "_i.json") {
			return nil
		}

		body, err := util.ReadFile(abs_path)

		if err != nil {
			return err
		}

		id := gjson.GetBytes(body, "photo.id")

		if !id.Exists() {
			return errors.New("Unabled to determine photo ID")
		}

		title := gjson.GetBytes(body, "photo.title._content")

		if !title.Exists() {
			return errors.New("Unabled to determine title")
		}

		secret := gjson.GetBytes(body, "photo.secret") // secret not originalsecret because we're going to display "_z.jpg"

		if !secret.Exists() {
			return errors.New("Unabled to determine title")
		}

		taken := gjson.GetBytes(body, "photo.dates.taken")

		if !taken.Exists() {
			return errors.New("Unabled to determine date taken")
		}

		dt, err := time.Parse("2006-01-02", taken.String())

		if err != nil {
			return err
		}

		item := StaticIndexItem{
			id:     id.Int(),
			title:  title.String(),
			secret: secret.String(),
			date:   dt,
		}

		mu.Lock()
		items = append(items, &item)
		mu.Unlock()

		return nil
	}

	err := filepath.Walk(r.Root, cb)

	if err != nil {
		return nil, err
	}

	count := len(items)
	breadcrumbs := make(map[int64][]int)

	for idx, item := range items {

		prev_idx := -1
		next_idx := -1

		id := item.ID()

		if idx != 0 {
			prev_idx = idx - 1
		}

		if idx < count-1 {
			next_idx = idx + 1
		}

		breadcrumbs[id] = []int{prev_idx, next_idx}
	}

	idx := StaticIndex{
		items:       items,
		breadcrumbs: breadcrumbs,
	}

	return &idx, err
}
*/
