package rules

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
)

var (
	nonAlphaNum     = regexp.MustCompile(`[^a-z0-9]+`)
	spaceCollapse   = regexp.MustCompile(`\s+`)
	abvPercentRe    = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*%`)
	proofRe         = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*proof`)
	netContentsRe   = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(ml|l|oz|fl\.?\s*oz)`)
	warningPrefixRe = regexp.MustCompile(`(?i)^\s*government\s+warning\s*:`)
)

const officialWarningPrefix = "GOVERNMENT WARNING:"

var officialWarningBody = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(
	"GOVERNMENT WARNING: (1) According to the Surgeon General, women should not drink alcoholic beverages during pregnancy because of the risk of birth defects. (2) Consumption of alcoholic beverages impairs your ability to drive a car or operate machinery, and may cause health problems.",
	officialWarningPrefix,
)))

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) Evaluate(app domain.ApplicationData, extracted domain.ExtractedFields) []domain.FieldResult {
	return []domain.FieldResult{
		e.compareBrand(app.BrandName, extracted),
		e.compareClassType(app.ClassType, extracted),
		e.compareABV(app.AlcoholContent, extracted),
		e.compareNetContents(app.NetContents, extracted),
		e.compareProducer(app.ProducerAddress, extracted),
		e.compareCountry(app.CountryOfOrigin, extracted),
		e.compareWarning(app.GovernmentWarning, extracted),
	}
}

func (e *Engine) OverallStatus(fields []domain.FieldResult) domain.OverallStatus {
	hasMismatch := false
	hasReview := false
	for _, field := range fields {
		switch field.Status {
		case domain.FieldMismatch:
			hasMismatch = true
		case domain.FieldNeedsReview:
			hasReview = true
		}
	}
	if hasMismatch {
		return domain.StatusFail
	}
	if hasReview {
		return domain.StatusNeedsReview
	}
	return domain.StatusPass
}

func (e *Engine) compareBrand(expected string, extracted domain.ExtractedFields) domain.FieldResult {
	found := extracted.BrandName
	status := domain.FieldMatch
	message := "Brand name matches after normalization."

	normExpected := normalizeBrand(expected)
	normFound := normalizeBrand(found)
	if normExpected == "" || normFound == "" {
		status = domain.FieldNeedsReview
		message = "Brand name could not be read clearly from the label."
	} else if normExpected != normFound && !fuzzyMatch(normExpected, normFound, 0.88) {
		status = domain.FieldMismatch
		message = "Brand name on the label does not match the application."
	}

	return domain.FieldResult{
		Field: "brand_name", Status: status, Expected: expected, Found: found,
		Evidence: extracted.Evidence["brand_name"], Message: message,
	}
}

func (e *Engine) compareClassType(expected string, extracted domain.ExtractedFields) domain.FieldResult {
	found := extracted.ClassType
	status := domain.FieldMatch
	message := "Class/type designation matches."

	if normalizeText(expected) == "" || normalizeText(found) == "" {
		status = domain.FieldNeedsReview
		message = "Class/type could not be confirmed from the label."
	} else if !containsFuzzy(normalizeText(expected), normalizeText(found)) {
		status = domain.FieldMismatch
		message = "Class/type on the label does not match the application."
	}

	return domain.FieldResult{
		Field: "class_type", Status: status, Expected: expected, Found: found,
		Evidence: extracted.Evidence["class_type"], Message: message,
	}
}

func (e *Engine) compareABV(expected string, extracted domain.ExtractedFields) domain.FieldResult {
	found := extracted.AlcoholContent
	expectedABV, okExpected := parseABV(expected)
	foundABV, okFound := parseABV(found)

	status := domain.FieldMatch
	message := "Alcohol content matches."

	switch {
	case !okExpected || !okFound:
		status = domain.FieldNeedsReview
		message = "Alcohol content could not be parsed reliably."
	case abs(expectedABV-foundABV) > 0.3:
		status = domain.FieldMismatch
		message = "Alcohol content on the label does not match the application."
	}

	return domain.FieldResult{
		Field: "alcohol_content", Status: status, Expected: expected, Found: found,
		Evidence: extracted.Evidence["alcohol_content"], Message: message,
	}
}

func (e *Engine) compareNetContents(expected string, extracted domain.ExtractedFields) domain.FieldResult {
	found := extracted.NetContents
	expectedValue, expectedUnit, okExpected := parseNetContents(expected)
	foundValue, foundUnit, okFound := parseNetContents(found)

	status := domain.FieldMatch
	message := "Net contents match."

	switch {
	case !okExpected || !okFound:
		status = domain.FieldNeedsReview
		message = "Net contents could not be parsed reliably."
	case expectedUnit != foundUnit || abs(expectedValue-foundValue) > 1:
		status = domain.FieldMismatch
		message = "Net contents on the label do not match the application."
	}

	return domain.FieldResult{
		Field: "net_contents", Status: status, Expected: expected, Found: found,
		Evidence: extracted.Evidence["net_contents"], Message: message,
	}
}

func (e *Engine) compareProducer(expected string, extracted domain.ExtractedFields) domain.FieldResult {
	found := extracted.ProducerAddress
	status := domain.FieldMatch
	message := "Producer/bottler address matches."

	if normalizeText(expected) == "" {
		status = domain.FieldNeedsReview
		message = "No producer address provided in the application."
	} else if normalizeText(found) == "" {
		status = domain.FieldNeedsReview
		message = "Producer/bottler address was not clearly visible on the label."
	} else if !containsFuzzy(normalizeText(expected), normalizeText(found)) {
		status = domain.FieldMismatch
		message = "Producer/bottler address does not match the application."
	}

	return domain.FieldResult{
		Field: "producer_address", Status: status, Expected: expected, Found: found,
		Evidence: extracted.Evidence["producer_address"], Message: message,
	}
}

func (e *Engine) compareCountry(expected string, extracted domain.ExtractedFields) domain.FieldResult {
	found := extracted.CountryOfOrigin
	status := domain.FieldMatch
	message := "Country of origin matches."

	if normalizeText(expected) == "" {
		status = domain.FieldNeedsReview
		message = "Country of origin was not provided in the application."
	} else if normalizeText(found) == "" {
		status = domain.FieldNeedsReview
		message = "Country of origin was not clearly visible on the label."
	} else if !containsFuzzy(normalizeText(expected), normalizeText(found)) {
		status = domain.FieldMismatch
		message = "Country of origin does not match the application."
	}

	return domain.FieldResult{
		Field: "country_of_origin", Status: status, Expected: expected, Found: found,
		Evidence: extracted.Evidence["country_of_origin"], Message: message,
	}
}

func (e *Engine) compareWarning(expected string, extracted domain.ExtractedFields) domain.FieldResult {
	found := extracted.GovernmentWarning
	status := domain.FieldMatch
	message := "Government warning statement matches required wording."

	if strings.TrimSpace(found) == "" {
		return domain.FieldResult{
			Field: "government_warning", Status: domain.FieldMismatch,
			Expected: expected, Found: found, Evidence: extracted.Evidence["government_warning"],
			Message: "Government warning statement is missing from the label.",
		}
	}

	if !warningPrefixRe.MatchString(found) || !strings.Contains(found, officialWarningPrefix) {
		status = domain.FieldMismatch
		message = "Government warning prefix must be GOVERNMENT WARNING: in all caps."
	} else if tokenOverlap(normalizeText(strings.TrimPrefix(found, officialWarningPrefix)), officialWarningBody) < 0.75 {
		status = domain.FieldMismatch
		message = "Government warning wording does not match the required statement."
	}

	if conf, ok := extracted.Confidence["government_warning"]; ok && conf < 0.7 && status == domain.FieldMatch {
		status = domain.FieldNeedsReview
		message = "Warning text was read with low confidence; bold formatting could not be verified automatically."
	}

	return domain.FieldResult{
		Field: "government_warning", Status: status, Expected: expected, Found: found,
		Evidence: extracted.Evidence["government_warning"], Message: message,
	}
}

func normalizeBrand(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "'", "")
	value = strings.ReplaceAll(value, "’", "")
	value = nonAlphaNum.ReplaceAllString(value, " ")
	return spaceCollapse.ReplaceAllString(strings.TrimSpace(value), " ")
}

func normalizeText(value string) string {
	value = strings.ToLower(value)
	value = nonAlphaNum.ReplaceAllString(value, " ")
	return spaceCollapse.ReplaceAllString(strings.TrimSpace(value), " ")
}

func fuzzyMatch(a, b string, threshold float64) bool {
	if a == b {
		return true
	}
	return tokenOverlap(a, b) >= threshold
}

func containsFuzzy(expected, found string) bool {
	if expected == found || strings.Contains(found, expected) || strings.Contains(expected, found) {
		return true
	}
	return tokenOverlap(expected, found) >= 0.8
}

func tokenOverlap(a, b string) float64 {
	aTokens := strings.Fields(a)
	bTokens := strings.Fields(b)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}

	bSet := make(map[string]struct{}, len(bTokens))
	for _, token := range bTokens {
		bSet[token] = struct{}{}
	}

	matches := 0
	for _, token := range aTokens {
		if _, ok := bSet[token]; ok {
			matches++
		}
	}

	denominator := len(aTokens)
	if len(bTokens) > denominator {
		denominator = len(bTokens)
	}
	return float64(matches) / float64(denominator)
}

func parseABV(value string) (float64, bool) {
	if match := abvPercentRe.FindStringSubmatch(value); len(match) == 2 {
		return parseFloat(match[1])
	}
	if match := proofRe.FindStringSubmatch(value); len(match) == 2 {
		proof, ok := parseFloat(match[1])
		if !ok {
			return 0, false
		}
		return proof / 2, true
	}
	return parseFloat(strings.TrimSpace(value))
}

func parseNetContents(value string) (float64, string, bool) {
	match := netContentsRe.FindStringSubmatch(value)
	if len(match) != 3 {
		return 0, "", false
	}
	amount, ok := parseFloat(match[1])
	if !ok {
		return 0, "", false
	}
	unit := strings.ToLower(strings.ReplaceAll(match[2], " ", ""))
	return amount, unit, true
}

func parseFloat(value string) (float64, bool) {
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsDigit(r) || r == '.' {
			builder.WriteRune(r)
		} else if builder.Len() > 0 {
			break
		}
	}
	if builder.Len() == 0 {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(builder.String(), 64)
	return parsed, err == nil
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
