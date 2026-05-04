package api

import (
	"encoding/json"
	"net/http"
)

const maxUploadize = 32 << 20

func compressHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadize); err != nil {
		jsonError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
	}
	defer r.MultipartForm.RemoveAll()
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
