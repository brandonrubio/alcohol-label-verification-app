package ai

import (
	"context"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
)

type Extractor interface {
	Extract(ctx context.Context, imageName string, imageBytes []byte, mimeType string) (domain.ExtractedFields, error)
}
