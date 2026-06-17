package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/config"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/verification"
)

type Handler struct {
	cfg     config.Config
	service *verification.Service
}

func NewHandler(cfg config.Config, service *verification.Service) *Handler {
	return &Handler{cfg: cfg, service: service}
}

func (h *Handler) Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		unauthorized(w, "")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) CreateVerification(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		unauthorized(w, "")
		return
	}

	app, input, err := h.parseVerificationForm(r)
	if err != nil {
		badRequest(w, err)
		return
	}

	result, err := h.service.VerifyOne(r.Context(), user.ID, app, input)
	if err != nil {
		writeError(w, http.StatusBadGateway, "verification failed")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) GetVerification(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		unauthorized(w, "")
		return
	}

	id := r.PathValue("id")
	if id == "" {
		notFound(w)
		return
	}

	result, err := h.service.GetVerification(r.Context(), user.ID, id)
	if err != nil {
		notFound(w)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) ListVerifications(w http.ResponseWriter, r *http.Request) {
	user, ok := userFromContext(r.Context())
	if !ok {
		unauthorized(w, "")
		return
	}

	results, err := h.service.ListVerifications(r.Context(), user.ID, 25)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list verifications")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"results": results})
}

func (h *Handler) parseVerificationForm(r *http.Request) (domain.ApplicationData, domain.LabelInput, error) {
	if err := r.ParseMultipartForm(h.cfg.MaxUploadBytes); err != nil {
		return domain.ApplicationData{}, domain.LabelInput{}, err
	}

	app, err := parseApplicationJSON(r.FormValue("application"))
	if err != nil {
		return domain.ApplicationData{}, domain.LabelInput{}, err
	}

	files := r.MultipartForm.File["image"]
	if len(files) == 0 {
		return domain.ApplicationData{}, domain.LabelInput{}, errors.New("image is required")
	}

	input, err := fileToLabelInput(files[0], h.cfg.MaxUploadBytes)
	if err != nil {
		return domain.ApplicationData{}, domain.LabelInput{}, err
	}

	return app, input, nil
}

func parseApplicationJSON(raw string) (domain.ApplicationData, error) {
	if strings.TrimSpace(raw) == "" {
		return domain.ApplicationData{}, errors.New("application data is required")
	}

	var app domain.ApplicationData
	if err := json.Unmarshal([]byte(raw), &app); err != nil {
		return domain.ApplicationData{}, errors.New("application must be valid JSON")
	}
	return app, nil
}

func fileToLabelInput(header *multipart.FileHeader, maxBytes int64) (domain.LabelInput, error) {
	file, err := header.Open()
	if err != nil {
		return domain.LabelInput{}, err
	}
	defer file.Close()

	limited := io.LimitReader(file, maxBytes+1)
	bytes, err := io.ReadAll(limited)
	if err != nil {
		return domain.LabelInput{}, err
	}
	if int64(len(bytes)) > maxBytes {
		return domain.LabelInput{}, errors.New("upload exceeds size limit")
	}

	return domain.LabelInput{
		ImageName:  header.Filename,
		ImageBytes: bytes,
		MimeType:   header.Header.Get("Content-Type"),
	}, nil
}
