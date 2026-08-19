package handler

import (
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/handler/maimai"
)

func MaimaiHandler(w http.ResponseWriter, r *http.Request) {
	maimai.MaimaiHandler(w, r)
}
