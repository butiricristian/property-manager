package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

type ExtractedExpense struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Date        string  `json:"date"`
}

var validCategories = map[string]bool{
	"electricity":    true,
	"heating":        true,
	"water":          true,
	"maintenance":    true,
	"cleaning":       true,
	"other":          true,
	"partner_payout": true,
}

type BillExtractService struct {
	client anthropic.Client
}

func NewBillExtractService(apiKey string) *BillExtractService {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	return &BillExtractService{client: client}
}

func (s *BillExtractService) ExtractFromPDF(ctx context.Context, pdfBytes []byte) (*ExtractedExpense, error) {
	encoded := base64.StdEncoding.EncodeToString(pdfBytes)

	pdfSource := anthropic.Base64PDFSourceParam{
		Data: encoded,
	}

	resp, err := s.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: 1024,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(
				anthropic.NewDocumentBlock(pdfSource),
				anthropic.NewTextBlock(`Extract the expense details from this bill/invoice PDF. Return a JSON object with these fields:
- "category": one of "electricity", "heating", "water", "maintenance", "cleaning", "other"
- "description": a short description of the bill (e.g. "Electricity bill January 2025")
- "amount": the total amount due as a number (no currency symbols)
- "currency": the 3-letter currency code (e.g. "EUR", "USD", "RON")
- "date": the bill/invoice date in YYYY-MM-DD format

Return ONLY the JSON object, no other text.`),
			),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude API call failed: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	text := resp.Content[0].Text
	text = stripMarkdownFences(text)

	var result ExtractedExpense
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Claude response: %w (raw: %s)", err, text)
	}

	if !validCategories[result.Category] {
		result.Category = "other"
	}

	return &result, nil
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if after, ok := strings.CutPrefix(s, "```json"); ok {
		s = after
	} else if after, ok := strings.CutPrefix(s, "```"); ok {
		s = after
	}
	if before, ok := strings.CutSuffix(s, "```"); ok {
		s = before
	}
	return strings.TrimSpace(s)
}
