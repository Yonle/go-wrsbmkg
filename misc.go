package wrsbmkg

import (
	"time"
)

var DEFAULT_API_URL string = "https://bmkg-content-inatews.storage.googleapis.com"

// Unix Milli Now
func umn() int64 {
	now := time.Now()
	return now.UnixMilli()
}
