package handler

import (
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/handler/ongeki"
)

func OngekiHandler(w http.ResponseWriter, r *http.Request) {
	ongeki.Handler(w, r)
}
