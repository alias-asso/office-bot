package main

import (
	"fmt"
	"net/http"
)

var mux *http.ServeMux

func init() {
	mux = http.NewServeMux()

	mux.HandleFunc("GET /", handleStatus)

	mux.HandleFunc("POST /", handleToggle)
}

func handleStatus(w http.ResponseWriter, req *http.Request) {
	status, err := GetStatusString()
	if err != nil {
		fmt.Fprint(w, "error: "+err.Error())
	}
	fmt.Fprint(w, status)
}

func handleToggle(w http.ResponseWriter, req *http.Request) {
	newStatus, err := ToggleStatus()
	if err != nil {
		fmt.Fprint(w, "error")
	}
	fmt.Fprint(w, "new status: "+newStatus)
}
