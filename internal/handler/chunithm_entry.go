package handler

import (
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/handler/chunithm"
)

func ChuniHandler(w http.ResponseWriter, r *http.Request) {
	chunithm.Handler(w, r)
}
