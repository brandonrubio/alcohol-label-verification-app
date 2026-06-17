package verification

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/ai"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/rules"
	"github.com/brandon/alcohol-label-verification-app/backend/internal/store"
)

type Service struct {
	extractor ai.Extractor
	rules     *rules.Engine
	repo      *store.Repository
}

func NewService(extractor ai.Extractor, engine *rules.Engine, repo *store.Repository) *Service {
	return &Service{
		extractor: extractor,
		rules:     engine,
		repo:      repo,
	}
}

func (s *Service) VerifyOne(
	ctx context.Context,
	userID string,
	app domain.ApplicationData,
	input domain.LabelInput,
) (domain.Result, error) {
	start := time.Now()

	mimeType := input.MimeType
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(input.ImageName))
	}
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	extracted, err := s.extractor.Extract(ctx, input.ImageName, input.ImageBytes, mimeType)
	if err != nil {
		return domain.Result{}, fmt.Errorf("extract label fields: %w", err)
	}

	fields := s.rules.Evaluate(app, extracted)
	status := s.rules.OverallStatus(fields)

	result := domain.Result{
		ID:           uuid.NewString(),
		UserID:       userID,
		Status:       status,
		ImageName:    input.ImageName,
		Application:  app,
		Extracted:    extracted,
		Fields:       fields,
		ProcessingMS: int(time.Since(start).Milliseconds()),
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateVerification(ctx, result); err != nil {
		return domain.Result{}, fmt.Errorf("persist verification: %w", err)
	}

	return result, nil
}

func (s *Service) GetVerification(ctx context.Context, userID, id string) (domain.Result, error) {
	return s.repo.GetVerification(ctx, userID, id)
}

func (s *Service) ListVerifications(ctx context.Context, userID string, limit int) ([]domain.Result, error) {
	return s.repo.ListVerifications(ctx, userID, limit)
}
