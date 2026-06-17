package verification

import (
	"context"
	"fmt"
	"mime"
	"path/filepath"
	"sync"
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
	batchSize int
}

func NewService(extractor ai.Extractor, engine *rules.Engine, repo *store.Repository, batchConcurrency int) *Service {
	if batchConcurrency <= 0 {
		batchConcurrency = 4
	}
	return &Service{
		extractor: extractor,
		rules:     engine,
		repo:      repo,
		batchSize: batchConcurrency,
	}
}

func (s *Service) VerifyOne(
	ctx context.Context,
	userID string,
	app domain.ApplicationData,
	input domain.LabelInput,
	batchID *string,
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
		BatchID:      batchID,
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

func (s *Service) VerifyBatch(
	ctx context.Context,
	userID string,
	app domain.ApplicationData,
	inputs []domain.LabelInput,
) (domain.Batch, error) {
	batch, err := s.repo.CreateBatch(ctx, userID, len(inputs))
	if err != nil {
		return domain.Batch{}, err
	}

	sem := make(chan struct{}, s.batchSize)
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make([]domain.Result, 0, len(inputs))
	errs := make([]error, 0)

	for _, input := range inputs {
		wg.Add(1)
		go func(input domain.LabelInput) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			batchID := batch.ID
			result, err := s.VerifyOne(ctx, userID, app, input, &batchID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
				return
			}
			results = append(results, result)
		}(input)
	}

	wg.Wait()

	status := "completed"
	if len(errs) > 0 && len(results) == 0 {
		status = "failed"
	} else if len(errs) > 0 {
		status = "completed_with_errors"
	}

	if err := s.repo.UpdateBatchProgress(ctx, batch.ID, len(results), status); err != nil {
		return domain.Batch{}, err
	}

	batch.Status = status
	batch.CompletedCount = len(results)
	batch.Results = results
	batch.UpdatedAt = time.Now().UTC()
	return batch, nil
}

func (s *Service) GetVerification(ctx context.Context, userID, id string) (domain.Result, error) {
	return s.repo.GetVerification(ctx, userID, id)
}

func (s *Service) ListVerifications(ctx context.Context, userID string, limit int) ([]domain.Result, error) {
	return s.repo.ListVerifications(ctx, userID, limit)
}

func (s *Service) GetBatch(ctx context.Context, userID, id string) (domain.Batch, error) {
	return s.repo.GetBatch(ctx, userID, id)
}
