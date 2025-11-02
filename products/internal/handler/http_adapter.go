package handler

import (
	"net/http"
)

type HTTPHandler interface {
	HandlerFunc() http.HandlerFunc
}

type GinAdapter struct {
	handler func(w http.ResponseWriter, r *http.Request)
}

func NewGinAdapter(handler func(w http.ResponseWriter, r *http.Request)) HTTPHandler {
	return &GinAdapter{handler: handler}
}

func (a *GinAdapter) HandlerFunc() http.HandlerFunc {
	return a.handler
}

