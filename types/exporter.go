package types

import (
	"net/url"
)

type Exporter struct {
	URL  *url.URL
	Path *string
}

func (e *Exporter) GetURL() *url.URL {
	return e.URL
}
