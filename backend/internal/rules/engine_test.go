package rules

import (
	"testing"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
)

func TestEngineBrandNormalization(t *testing.T) {
	engine := NewEngine()
	app := domain.ApplicationData{BrandName: "Stone's Throw"}
	extracted := domain.ExtractedFields{
		BrandName:  "STONE'S THROW",
		Evidence:   map[string]string{"brand_name": "STONE'S THROW"},
		Confidence: map[string]float64{"brand_name": 0.95},
	}

	results := engine.Evaluate(app, extracted)
	if results[0].Status != domain.FieldMatch {
		t.Fatalf("expected brand match, got %s", results[0].Status)
	}
}

func TestEngineABVMismatch(t *testing.T) {
	engine := NewEngine()
	app := domain.ApplicationData{
		BrandName:      "OLD TOM DISTILLERY",
		AlcoholContent: "45% Alc./Vol.",
	}
	extracted := domain.ExtractedFields{
		BrandName:      "OLD TOM DISTILLERY",
		AlcoholContent: "40% Alc./Vol.",
		Evidence:       map[string]string{},
		Confidence:     map[string]float64{"alcohol_content": 0.9},
	}

	results := engine.Evaluate(app, extracted)
	for _, result := range results {
		if result.Field == "alcohol_content" && result.Status != domain.FieldMismatch {
			t.Fatalf("expected alcohol mismatch, got %s", result.Status)
		}
	}
}

func TestEngineWarningPrefix(t *testing.T) {
	engine := NewEngine()
	app := domain.ApplicationData{
		GovernmentWarning: "GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.",
	}
	extracted := domain.ExtractedFields{
		GovernmentWarning: "Government Warning: harmful",
		Evidence:          map[string]string{},
		Confidence:        map[string]float64{"government_warning": 0.9},
	}

	results := engine.Evaluate(app, extracted)
	for _, result := range results {
		if result.Field == "government_warning" && result.Status != domain.FieldMismatch {
			t.Fatalf("expected warning mismatch, got %s", result.Status)
		}
	}
}

func TestOverallStatusFailOnMismatch(t *testing.T) {
	engine := NewEngine()
	status := engine.OverallStatus([]domain.FieldResult{
		{Status: domain.FieldMatch},
		{Status: domain.FieldMismatch},
	})
	if status != domain.StatusFail {
		t.Fatalf("expected fail, got %s", status)
	}
}
