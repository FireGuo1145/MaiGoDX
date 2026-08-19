package handler

import (
	"net/http"

	"github.com/FireGuo1145/MaiGoDX/internal/handler/aime"
)

func StartAimeDB() { aime.StartAimeDB() }

func HandleAllNetSelfTest(w http.ResponseWriter, r *http.Request) {
	aime.HandleAllNetSelfTest(w, r)
}

func HandleNaomiTest(w http.ResponseWriter, r *http.Request) {
	aime.HandleNaomiTest(w, r)
}

func HandleAllNetPowerOn(w http.ResponseWriter, r *http.Request) {
	aime.HandleAllNetPowerOn(w, r)
}

func HandleAllNetDownloadOrder(w http.ResponseWriter, r *http.Request) {
	aime.HandleAllNetDownloadOrder(w, r)
}

func HandleBillingRequest(w http.ResponseWriter, r *http.Request) {
	aime.HandleBillingRequest(w, r)
}

func HandleTerminalMaimai(w http.ResponseWriter, r *http.Request) {
	aime.HandleTerminalMaimai(w, r)
}
