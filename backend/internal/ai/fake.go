package ai

import (
	"context"
	"strings"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
)

type FakeExtractor struct{}

func NewFakeExtractor() *FakeExtractor {
	return &FakeExtractor{}
}

func (f *FakeExtractor) Extract(
	_ context.Context,
	imageName string,
	_ []byte,
	_ string,
) (domain.ExtractedFields, error) {
	lower := strings.ToLower(imageName)

	fields := domain.ExtractedFields{
		BrandName:         "OLD TOM DISTILLERY",
		ClassType:         "Kentucky Straight Bourbon Whiskey",
		AlcoholContent:    "45% Alc./Vol. (90 Proof)",
		NetContents:       "750 mL",
		ProducerAddress:   "Old Tom Distillery, Louisville, KY",
		CountryOfOrigin:   "United States",
		GovernmentWarning: "GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.",
		Evidence: map[string]string{
			"brand_name":         "OLD TOM DISTILLERY",
			"government_warning": "GOVERNMENT WARNING:",
		},
		Confidence: map[string]float64{
			"brand_name":         0.95,
			"class_type":         0.9,
			"alcohol_content":    0.92,
			"net_contents":       0.9,
			"producer_address":   0.75,
			"country_of_origin":  0.8,
			"government_warning": 0.88,
		},
	}

	if strings.Contains(lower, "mismatch") {
		fields.BrandName = "WRONG BRAND"
		fields.AlcoholContent = "40% Alc./Vol."
	}

	if strings.Contains(lower, "warning") {
		fields.GovernmentWarning = "Government Warning: consumption may be harmful"
		fields.Confidence["government_warning"] = 0.55
	}

	return fields, nil
}
