package domain

import "time"

type ApplicationData struct {
	BrandName         string `json:"brand_name"`
	ClassType         string `json:"class_type"`
	AlcoholContent    string `json:"alcohol_content"`
	NetContents       string `json:"net_contents"`
	ProducerAddress   string `json:"producer_address"`
	CountryOfOrigin   string `json:"country_of_origin"`
	GovernmentWarning string `json:"government_warning"`
}

type ExtractedFields struct {
	BrandName         string             `json:"brand_name"`
	ClassType         string             `json:"class_type"`
	AlcoholContent    string             `json:"alcohol_content"`
	NetContents       string             `json:"net_contents"`
	ProducerAddress   string             `json:"producer_address"`
	CountryOfOrigin   string             `json:"country_of_origin"`
	GovernmentWarning string             `json:"government_warning"`
	Evidence          map[string]string  `json:"evidence"`
	Confidence        map[string]float64 `json:"confidence"`
}

type FieldStatus string

const (
	FieldMatch       FieldStatus = "match"
	FieldMismatch    FieldStatus = "mismatch"
	FieldNeedsReview FieldStatus = "needs_review"
)

type OverallStatus string

const (
	StatusPass        OverallStatus = "pass"
	StatusNeedsReview OverallStatus = "needs_review"
	StatusFail        OverallStatus = "fail"
)

type FieldResult struct {
	Field    string      `json:"field"`
	Status   FieldStatus `json:"status"`
	Expected string      `json:"expected"`
	Found    string      `json:"found"`
	Evidence string      `json:"evidence,omitempty"`
	Message  string      `json:"message,omitempty"`
}

type Result struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id,omitempty"`
	BatchID      *string         `json:"batch_id,omitempty"`
	Status       OverallStatus   `json:"status"`
	ImageName    string          `json:"image_name"`
	Application  ApplicationData `json:"application"`
	Extracted    ExtractedFields `json:"extracted"`
	Fields       []FieldResult   `json:"fields"`
	ProcessingMS int             `json:"processing_ms"`
	CreatedAt    time.Time       `json:"created_at"`
}

type Batch struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id,omitempty"`
	Status         string    `json:"status"`
	TotalCount     int       `json:"total_count"`
	CompletedCount int       `json:"completed_count"`
	Results        []Result  `json:"results,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type LabelInput struct {
	ImageName  string
	ImageBytes []byte
	MimeType   string
}
