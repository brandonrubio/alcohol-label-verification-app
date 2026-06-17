package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/brandon/alcohol-label-verification-app/backend/internal/domain"
)

type GeminiExtractor struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGeminiExtractor(apiKey, model string, timeout time.Duration) *GeminiExtractor {
	return &GeminiExtractor{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: timeout},
	}
}

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string        `json:"text,omitempty"`
	InlineData *geminiInline `json:"inlineData,omitempty"`
}

type geminiInline struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiGenerationConfig struct {
	ResponseMIMEType string `json:"responseMimeType"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (g *GeminiExtractor) Extract(
	ctx context.Context,
	imageName string,
	imageBytes []byte,
	mimeType string,
) (domain.ExtractedFields, error) {
	if mimeType == "" {
		mimeType = "image/jpeg"
	}

	prompt := `You are extracting alcohol label fields for TTB compliance review.
Return JSON only with these keys:
brand_name, class_type, alcohol_content, net_contents, producer_address, country_of_origin, government_warning,
evidence (object mapping field names to short quoted snippets from the label),
confidence (object mapping field names to 0-1 confidence scores).
If a field is not visible, use an empty string and low confidence.`

	reqBody := geminiRequest{
		Contents: []geminiContent{{
			Parts: []geminiPart{
				{Text: prompt},
				{
					InlineData: &geminiInline{
						MimeType: mimeType,
						Data:     base64.StdEncoding.EncodeToString(imageBytes),
					},
				},
				{Text: fmt.Sprintf("Image filename: %s", imageName)},
			},
		}},
		GenerationConfig: geminiGenerationConfig{
			ResponseMIMEType: "application/json",
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return domain.ExtractedFields{}, fmt.Errorf("marshal gemini request: %w", err)
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.model,
		g.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return domain.ExtractedFields{}, fmt.Errorf("create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return domain.ExtractedFields{}, fmt.Errorf("call gemini: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ExtractedFields{}, fmt.Errorf("read gemini response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return domain.ExtractedFields{}, fmt.Errorf("gemini status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return domain.ExtractedFields{}, fmt.Errorf("decode gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return domain.ExtractedFields{}, fmt.Errorf("gemini returned no candidates")
	}

	text := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	var extracted domain.ExtractedFields
	if err := json.Unmarshal([]byte(text), &extracted); err != nil {
		return domain.ExtractedFields{}, fmt.Errorf("decode gemini json payload: %w", err)
	}

	if extracted.Evidence == nil {
		extracted.Evidence = map[string]string{}
	}
	if extracted.Confidence == nil {
		extracted.Confidence = map[string]float64{}
	}

	return extracted, nil
}
