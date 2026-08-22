package ask

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/devkrishna/laterbase/backend/internal/db"
	"github.com/devkrishna/laterbase/backend/internal/models"
)

const (
	apiURL          = "https://api.openai.com/v1/chat/completions"
	maxCandidates   = 400  // cap candidate pool sent to the ranker
	rankTopN        = 8    // max links cited per ask
	contentTruncate = 8000 // chars of content_text per link when full-text grounding is on
)

// Known LLM pricing per model ($ per 1M tokens). Mirrors the worker table;
// update the gpt-5.6-luna entry when its pricing is known (falls back to
// gpt-4o-mini until then).
var llmPricing = map[string]struct{ input, output float64 }{
	"gpt-4o-mini": {0.15, 0.60},
	"gpt-4o":      {2.50, 10.00},
}

func computeCost(model string, inputTokens, outputTokens int) float64 {
	p, ok := llmPricing[model]
	if !ok {
		p = llmPricing["gpt-4o-mini"]
	}
	return (float64(inputTokens) * p.input / 1_000_000) + (float64(outputTokens) * p.output / 1_000_000)
}

type rankedMatch struct {
	ID     int    `json:"id"`
	Reason string `json:"reason"`
}

type rankResponse struct {
	Matches []rankedMatch `json:"matches"`
}

type usage struct {
	PromptTokens     int
	CompletionTokens int
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Run executes the two-stage ask pipeline:
//
//  1. Rank all ready links against the question using title+summary only.
//  2. Stream an answer grounded in the matched links' metadata (plus
//     content_text when includeFullText is set), emitting deltas via onDelta.
//
// onCitations (if non-nil) is invoked with the matched links immediately
// after ranking completes — before any answer deltas — so callers can
// surface results while the answer is still generating.
//
// Returns nil result without error when nothing in the archive matches —
// callers should distinguish that via Answer == "" && Citations == nil.
func Run(ctx context.Context, pool *pgxpool.Pool, apiKey, model, question string, includeFullText bool, onCitations func([]models.AskCitation), onDelta func(string)) (*models.AskResult, error) {
	candidates, err := db.ListAskCandidates(ctx, pool)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, nil
	}
	if len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}

	matches, rankUsage, err := rank(ctx, apiKey, model, question, candidates)
	if err != nil {
		return nil, fmt.Errorf("ranking failed: %w", err)
	}

	result := &models.AskResult{
		InputTokens: rankUsage.PromptTokens,
		Model:       model,
	}
	if len(matches) == 0 {
		result.OutputTokens = rankUsage.CompletionTokens
		return result, nil
	}

	byIndex := make(map[int]models.AskCandidate, len(candidates))
	for i, c := range candidates {
		byIndex[i+1] = c // prompts are 1-indexed
	}

	citations := make([]models.AskCitation, 0, len(matches))
	for _, m := range matches {
		c, ok := byIndex[m.ID]
		if !ok {
			continue
		}
		reason := strings.TrimSpace(m.Reason)
		if len(reason) > 200 {
			reason = reason[:200]
		}
		citations = append(citations, models.AskCitation{
			LinkID:    c.ID,
			Reason:    reason,
			URL:       c.URL,
			Title:     c.Title,
			CreatedAt: c.CreatedAt,
			GroupIDs:  c.GroupIDs,
		})
	}
	if len(citations) == 0 {
		return result, nil
	}
	result.Citations = citations

	if onCitations != nil {
		onCitations(citations)
	}

	answer, answerUsage, err := streamAnswer(ctx, pool, apiKey, model, question, candidates, citations, includeFullText, onDelta)
	if err != nil {
		return nil, fmt.Errorf("answer failed: %w", err)
	}

	result.Answer = answer
	result.InputTokens += answerUsage.PromptTokens
	result.OutputTokens = rankUsage.CompletionTokens + answerUsage.CompletionTokens

	if result.InputTokens > 0 || result.OutputTokens > 0 {
		cost := computeCost(model, result.InputTokens, result.OutputTokens)
		if err := db.InsertAskCost(ctx, pool, question, result.InputTokens, result.OutputTokens, cost, model); err != nil {
			log.Printf("Failed to insert ask cost: %v", err)
		}
	}

	return result, nil
}

// stage-1: rank candidates by title+summary relevance.
func rank(ctx context.Context, apiKey, model, question string, candidates []models.AskCandidate) ([]rankedMatch, *usage, error) {
	var sb strings.Builder
	sb.WriteString("Saved links:\n")
	for i, c := range candidates {
		summary := c.Summary
		if len(summary) > 300 {
			summary = summary[:300]
		}
		sb.WriteString(fmt.Sprintf("[%d] %s — %s\n", i+1, c.Title, summary))
	}

	systemPrompt := `You are a search ranker over a personal read-later link archive. Given a question and a numbered list of saved links (title — summary), return ONLY valid JSON (no markdown, no code fences):
{"matches": [{"id": <number>, "reason": "<one line, max ~15 words, why this link is relevant>"}]}

Rules:
- Include at most 8 matches, most relevant first
- Only include links genuinely relevant to the question
- If nothing is relevant, return {"matches": []}`

	body := map[string]interface{}{
		"model": model,
		"messages": []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Question: " + question + "\n\n" + sb.String()},
		},
		"max_completion_tokens": 800,
	}

	respBody, respUsage, err := chatCompletion(ctx, apiKey, body)
	if err != nil {
		return nil, nil, err
	}

	var parsed rankResponse
	if err := json.Unmarshal([]byte(stripFences(respBody)), &parsed); err != nil {
		return nil, &respUsage, fmt.Errorf("invalid ranker response: %w", err)
	}

	matches := []rankedMatch{}
	for _, m := range parsed.Matches {
		if m.ID >= 1 && m.ID <= len(candidates) && len(matches) < rankTopN {
			matches = append(matches, m)
		}
	}
	return matches, &respUsage, nil
}

// stage-2: stream a grounded answer over the cited links.
func streamAnswer(ctx context.Context, pool *pgxpool.Pool, apiKey, model, question string, candidates []models.AskCandidate, citations []models.AskCitation, includeFullText bool, onDelta func(string)) (string, *usage, error) {
	idToCandidate := make(map[uuid.UUID]models.AskCandidate, len(candidates))
	for _, c := range candidates {
		idToCandidate[c.ID] = c
	}

	var ctxBlocks strings.Builder
	for i, cit := range citations {
		c, ok := idToCandidate[cit.LinkID]
		if !ok {
			continue
		}
		fmt.Fprintf(&ctxBlocks, "[%d] %s\nURL: %s\nSummary: %s\n", i+1, c.Title, c.URL, c.Summary)
		if includeFullText {
			content, err := db.GetLinkContentText(ctx, nil, cit.LinkID)
			if err != nil {
				log.Printf("Failed to load content for citation %d: %v", i+1, err)
			} else if content != "" {
				if len(content) > contentTruncate {
					content = content[:contentTruncate]
				}
				fmt.Fprintf(&ctxBlocks, "Content: %s\n", content)
			}
		}
		ctxBlocks.WriteString("\n")
	}

	systemPrompt := `You answer questions about someone's personal saved-link archive. You are given the user's question and excerpts from the relevant saved links. Rules:
- Answer using ONLY the provided link material — never fabricate facts
- Cite sources inline as [1], [2] matching the numbered links
- If the saved links don't contain enough information to answer, say so plainly instead of guessing
- Be concise: 2-5 sentences unless the question needs more`

	body := map[string]interface{}{
		"model": model,
		"messages": []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: "Question: " + question + "\n\nRelevant saved links:\n\n" + ctxBlocks.String()},
		},
		"max_completion_tokens": 1000,
		"stream":                true,
		"stream_options":        map[string]bool{"include_usage": true},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", nil, fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(b))
	}

	var full strings.Builder
	u := &usage{}
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Usage *struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		for _, choice := range chunk.Choices {
			if choice.Delta.Content != "" {
				full.WriteString(choice.Delta.Content)
				if onDelta != nil {
					onDelta(choice.Delta.Content)
				}
			}
		}
		if chunk.Usage != nil {
			u.PromptTokens = chunk.Usage.PromptTokens
			u.CompletionTokens = chunk.Usage.CompletionTokens
		}
	}
	if err := scanner.Err(); err != nil {
		return full.String(), u, err
	}

	return full.String(), u, nil
}

// chatCompletion performs a single non-streaming completion and returns the
// message content plus token usage.
func chatCompletion(ctx context.Context, apiKey string, body map[string]interface{}) (string, usage, error) {
	var u usage
	payload, err := json.Marshal(body)
	if err != nil {
		return "", u, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(payload))
	if err != nil {
		return "", u, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", u, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", u, fmt.Errorf("LLM API error %d: %s", resp.StatusCode, string(b))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", u, err
	}
	u.PromptTokens = parsed.Usage.PromptTokens
	u.CompletionTokens = parsed.Usage.CompletionTokens

	if len(parsed.Choices) == 0 {
		return "", u, fmt.Errorf("no choices in LLM response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), u, nil
}

func stripFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
