package objects

import "time"

type Sale struct {
	Slug       string
	PlayerSlug string
	FirstName  string
	LastName   string
	Season     int64
	Price      uint64
	Date       time.Time
}
