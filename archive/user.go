package archive

import (
	"time"
)

type User struct {
	Username   string
	NSID       string
	FirstPhoto time.Time	// please rename me...
}
