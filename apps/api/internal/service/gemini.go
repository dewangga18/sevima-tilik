package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

type GeminiClient struct {
	client  *http.Client
	baseURL string
}

func NewGeminiClient() *GeminiClient {
	return &GeminiClient{client: &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse }}, baseURL: "https://generativelanguage.googleapis.com/v1beta/models/"}
}
func (g *GeminiClient) Generate(ctx context.Context, key, model, prompt string, schema map[string]any) (json.RawMessage, error) {
	payload := map[string]any{
		"systemInstruction": map[string]any{"parts": []any{map[string]string{"text": "Buat satu draft soal matematika kelas 4 dalam Bahasa Indonesia. Hanya ikuti spesifikasi skill. Jangan mengarang sumber atau klaim validasi. Output akan ditinjau manusia. Jawab JSON sesuai schema."}}},
		"contents":          []any{map[string]any{"role": "user", "parts": []any{map[string]string{"text": prompt}}}},
		"generationConfig":  map[string]any{"responseMimeType": "application/json", "responseJsonSchema": schema, "maxOutputTokens": 4096},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, ErrAIInput
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+model+":generateContent", bytes.NewReader(body))
	if err != nil {
		return nil, ErrAIProvider
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", key)
	res, err := g.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, ErrAIQuota
	}
	defer res.Body.Close()
	if res.StatusCode == 429 || res.StatusCode >= 500 {
		return nil, ErrAIQuota
	}
	if res.StatusCode != 200 {
		return nil, ErrAIProvider
	}
	raw, err := io.ReadAll(io.LimitReader(res.Body, (64<<10)+1))
	if err != nil || len(raw) > 64<<10 {
		return nil, ErrAIOutput
	}
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text    string `json:"text"`
					Thought bool   `json:"thought"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
	}
	if json.Unmarshal(raw, &response) != nil || len(response.Candidates) != 1 || response.Candidates[0].FinishReason != "STOP" {
		return nil, ErrAIOutput
	}
	var text strings.Builder
	for _, part := range response.Candidates[0].Content.Parts {
		if !part.Thought {
			text.WriteString(part.Text)
		}
	}
	result := json.RawMessage(text.String())
	if !json.Valid(result) || len(result) == 0 {
		return nil, ErrAIOutput
	}
	return result, nil
}
