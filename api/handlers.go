package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/prathmesh-d-glitch/go-zipper/archive"
)

// from 32 MB to 128 MB in-memory portion of multipart parsing
const maxUploadize = 128 << 20

//session store
//decompressed sessions held in memory so that individual files can be downloaded

type session struct {
	files []archive.ExtractedFile
}

var (
	sessionMU sync.RWMutex
	sessions  = make(map[string]*session)
	sessionID uint64
)

func newSessionID() string {
	sessionMU.Lock()
	defer sessionMU.Unlock()
	sessionID++
	return fmt.Sprintf("%d", sessionID)
}
func compressHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadize); err != nil {
		jsonError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
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

	if len(paths) == 0 {
		jsonError(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	archiveData, err := archive.CompressFiles(paths)
	if err != nil {
		jsonError(w, "compression failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="archive.fzp"`)
	w.Header().Set("Content-Length", strconv.Itoa(len(archiveData)))
	w.Write(archiveData)

}

func decompressHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxUploadize); err != nil {
		jsonError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, _, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "no archive file provided (use form field \"file\")", http.StatusBadRequest)
		return
	}
	defer file.Close()

	archiveBytes, err := io.ReadAll(file)
	if err != nil {
		jsonError(w, "failed to read uploaded archive: "+err.Error(), http.StatusBadRequest)
		return
	}

	extracted, err := archive.ExtractFiles(archiveBytes)
	if err != nil {
		jsonError(w, "decompression failed: "+err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if len(extracted) == 0 {
		jsonError(w, "archive contains no files", http.StatusUnprocessableEntity)
		return
	}

	// Single file → stream the raw bytes back directly.
	if len(extracted) == 1 {
		f := extracted[0]
		cd := fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(f.Name))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", cd)
		w.Header().Set("Content-Length", strconv.Itoa(len(f.Data)))
		w.WriteHeader(http.StatusOK)
		w.Write(f.Data)
		return
	}

	// Multiple files → store in session, return a manifest.
	sid := newSessionID()
	sess := &session{files: extracted}
	sessionMU.Lock()
	sessions[sid] = sess
	sessionMU.Unlock()

	names := make([]string, len(extracted))
	for i, f := range extracted {
		names[i] = f.Name
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"session_id": sid,
		"files":      names,
		"download_url_template": fmt.Sprintf(
			"GET /decompress/%s/{filename}", sid),
	})
}

func decompressFileHandler(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sessionID")
	name := chi.URLParam(r, "filename")

	sessionMU.RLock()
	sess, ok := sessions[sid]
	sessionMU.RUnlock()

	if !ok {
		jsonError(w, "session not found or expired", http.StatusNotFound)
		return
	}

	for _, f := range sess.files {
		if filepath.Base(f.Name) == name {
			cd := fmt.Sprintf(`attachment; filename="%s"`, name)
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", cd)
			w.Header().Set("Content-Length", strconv.Itoa(len(f.Data)))
			w.WriteHeader(http.StatusOK)
			w.Write(f.Data)
			return
		}
	}

	jsonError(w, fmt.Sprintf("file %q not found in session", name), http.StatusNotFound)
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
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
