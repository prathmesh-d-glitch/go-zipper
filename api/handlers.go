package api

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const maxUploadize = 32 << 20

func compressHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadize); err != nil {
		jsonError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
	}
	defer r.MultipartForm.RemoveAll()

	tmpDir, err := os.MkdirTemp("", "fz-compress-*")
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)

	var paths []string
	for _, fileHeaders := range r.MultipartForm.File {
		for _, fh := range fileHeaders {
			saved, err := saveUpload(fh, tmpDir)
			if err != nil {
				jsonError(w, "failed to process uploaded file: "+err.Error(), http.StatusBadRequest)
				return
			}
			paths = append(paths, saved)
		}
	}

}

func saveUpload(fh *multipart.FileHeader, dir string) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	dst := filepath.Join(dir, filepath.Base(fh.Filename))
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return "", err
	}

	return dst, nil
}

func jsonError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
