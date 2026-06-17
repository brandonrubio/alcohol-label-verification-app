package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
)

type Repository struct {
	db *DB
}

func NewRepository(db *DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateBatch(ctx context.Context, userID string, total int) (domain.Batch, error) {
	batch := domain.Batch{
		ID:             uuid.NewString(),
		UserID:         userID,
		Status:         "processing",
		TotalCount:     total,
		CompletedCount: 0,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO verification_batches (id, user_id, status, total_count, completed_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, batch.ID, userID, batch.Status, batch.TotalCount, batch.CompletedCount, batch.CreatedAt, batch.UpdatedAt)
	if err != nil {
		return domain.Batch{}, fmt.Errorf("insert batch: %w", err)
	}

	return batch, nil
}

func (r *Repository) UpdateBatchProgress(ctx context.Context, batchID string, completed int, status string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE verification_batches
		SET completed_count = $1, status = $2, updated_at = NOW()
		WHERE id = $3
	`, completed, status, batchID)
	if err != nil {
		return fmt.Errorf("update batch progress: %w", err)
	}
	return nil
}

func (r *Repository) CreateVerification(ctx context.Context, result domain.Result) error {
	appJSON, err := json.Marshal(result.Application)
	if err != nil {
		return fmt.Errorf("marshal application: %w", err)
	}
	extractedJSON, err := json.Marshal(result.Extracted)
	if err != nil {
		return fmt.Errorf("marshal extracted: %w", err)
	}
	fieldsJSON, err := json.Marshal(result.Fields)
	if err != nil {
		return fmt.Errorf("marshal fields: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO verifications (
			id, batch_id, user_id, status, image_name, application_json,
			extracted_json, field_results_json, processing_ms, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,
		result.ID,
		nullString(result.BatchID),
		result.UserID,
		string(result.Status),
		result.ImageName,
		appJSON,
		extractedJSON,
		fieldsJSON,
		result.ProcessingMS,
		result.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert verification: %w", err)
	}
	return nil
}

func (r *Repository) GetVerification(ctx context.Context, userID, id string) (domain.Result, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, batch_id, user_id, status, image_name, application_json,
		       extracted_json, field_results_json, processing_ms, created_at
		FROM verifications
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	return scanVerification(row)
}

func (r *Repository) ListVerifications(ctx context.Context, userID string, limit int) ([]domain.Result, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, batch_id, user_id, status, image_name, application_json,
		       extracted_json, field_results_json, processing_ms, created_at
		FROM verifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list verifications: %w", err)
	}
	defer rows.Close()

	results := make([]domain.Result, 0)
	for rows.Next() {
		result, err := scanVerification(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate verifications: %w", err)
	}
	return results, nil
}

func (r *Repository) GetBatch(ctx context.Context, userID, batchID string) (domain.Batch, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, status, total_count, completed_count, created_at, updated_at
		FROM verification_batches
		WHERE id = $1 AND user_id = $2
	`, batchID, userID)

	var batch domain.Batch
	if err := row.Scan(
		&batch.ID,
		&batch.UserID,
		&batch.Status,
		&batch.TotalCount,
		&batch.CompletedCount,
		&batch.CreatedAt,
		&batch.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Batch{}, fmt.Errorf("batch not found")
		}
		return domain.Batch{}, fmt.Errorf("get batch: %w", err)
	}

	results, err := r.listVerificationsByBatch(ctx, userID, batchID)
	if err != nil {
		return domain.Batch{}, err
	}
	batch.Results = results
	return batch, nil
}

func (r *Repository) listVerificationsByBatch(ctx context.Context, userID, batchID string) ([]domain.Result, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, batch_id, user_id, status, image_name, application_json,
		       extracted_json, field_results_json, processing_ms, created_at
		FROM verifications
		WHERE user_id = $1 AND batch_id = $2
		ORDER BY created_at ASC
	`, userID, batchID)
	if err != nil {
		return nil, fmt.Errorf("list batch verifications: %w", err)
	}
	defer rows.Close()

	results := make([]domain.Result, 0)
	for rows.Next() {
		result, err := scanVerification(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate batch verifications: %w", err)
	}
	return results, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanVerification(row rowScanner) (domain.Result, error) {
	var result domain.Result
	var batchID sql.NullString
	var appJSON, extractedJSON, fieldsJSON []byte

	if err := row.Scan(
		&result.ID,
		&batchID,
		&result.UserID,
		&result.Status,
		&result.ImageName,
		&appJSON,
		&extractedJSON,
		&fieldsJSON,
		&result.ProcessingMS,
		&result.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Result{}, fmt.Errorf("verification not found")
		}
		return domain.Result{}, fmt.Errorf("scan verification: %w", err)
	}

	if batchID.Valid {
		value := batchID.String
		result.BatchID = &value
	}

	if err := json.Unmarshal(appJSON, &result.Application); err != nil {
		return domain.Result{}, fmt.Errorf("decode application: %w", err)
	}
	if len(extractedJSON) > 0 {
		if err := json.Unmarshal(extractedJSON, &result.Extracted); err != nil {
			return domain.Result{}, fmt.Errorf("decode extracted: %w", err)
		}
	}
	if err := json.Unmarshal(fieldsJSON, &result.Fields); err != nil {
		return domain.Result{}, fmt.Errorf("decode fields: %w", err)
	}

	return result, nil
}

func nullString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
