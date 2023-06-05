package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/vkuksa/twatter/internal"
)

func renderErrorResponse(w http.ResponseWriter, r *http.Request, msg string, err error) {
	status := http.StatusInternalServerError

	var ierr *internal.Error
	if !errors.As(err, &ierr) {
		msg = "internal error"
	} else {
		switch ierr.Code() {
		case internal.ErrorCodeNotFound:
			status = http.StatusNotFound
		case internal.ErrorCodeInvalidArgument:
			status = http.StatusBadRequest
		case internal.ErrorCodeUnknown:
			fallthrough
		default:
			status = http.StatusInternalServerError
		}
	}

	render.Status(r, status)
	render.HTML(w, r, msg)
}
