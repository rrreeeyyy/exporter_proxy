package handler

import (
	"net/http"
	"net/url"
)

type ExporterHandler struct {
	URL *url.URL
}

func (handler *ExporterHandler) CreateDirector() func(req *http.Request) {
	return func(req *http.Request) {
		req.URL = handler.URL
	}
}
