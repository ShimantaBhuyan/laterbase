package categorizer

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-resty/resty/v2"
)

// CategorizeResult holds the result of LLM-based categorization.
type CategorizeResult struct {
	SourceType      string `json:"source_type"`
	SuggestedGroup  string `json:"suggested_group"`
	Summary         string `json:"summary"`
	Title           string `json:"title"`
	InputTokens     int     `json:"input_tokens"`
	OutputTokens    int     `json:"output_tokens"`
	ModelUsed       string  `json:"model_used"`
}

// Categorize uses the LLM to categorize a URL and its content.
// allowedGroups is the list of group names the suggestion must come from;
// "uncategorized" is always allowed as a last resort.
func Categorize(ctx context.Context, apiKey, model, url, content string, allowedGroups []string) (*CategorizeResult, error) {
	if apiKey == "" {
		return fallbackCategorize(url, content), nil
	}

	// Truncate content to 12000 chars
	if len(content) > 12000 {
		content = content[:12000]
	}

	groupsList := allowedGroups
	if len(groupsList) == 0 {
		groupsList = []string{"articles", "github", "docs", "videos", "tools"}
	}
	hasUncategorized := false
	for _, g := range groupsList {
		if strings.EqualFold(g, "uncategorized") {
			hasUncategorized = true
			break
		}
	}
	if !hasUncategorized {
		groupsList = append(groupsList, "uncategorized")
	}
	for i, g := range groupsList {
		groupsList[i] = strings.TrimSpace(g)
	}

	systemPrompt := `You are a link categorization assistant. Given a URL and the extracted page content, return ONLY valid JSON (no markdown, no code fences):
{
  "source_type": "article|github|youtube|twitter|docs|tools|other",
  "suggested_group": one of [` + quoteJoin(groupsList) + `],
  "summary": "2-3 sentence factual summary of what this page contains",
  "title": null or "cleaned title if the extracted one is missing or clearly wrong"
}

Rules:
- Infer source_type from URL domain and content patterns
- suggested_group MUST be exactly one of the listed group names — pick whichever best matches the page's TOPIC (what it is about), not just its format (what kind of site it is); do not invent new groups
- Format groups: articles=blog/news posts, github=code repos, docs=documentation sites, videos, tools=SaaS/dev utilities. A topic-specific group name in the list (anything else) wins over these whenever the page is primarily ABOUT that topic — e.g. an AI agents framework is about AI, so use an AI-named group rather than "tools"
- Look at the meta description, headings, and recurring terms in the content to determine the topic; also consider hints in the domain name (".ai", "ml", etc.)
- Use "uncategorized" ONLY when no listed group plausibly relates to the page
- If content is empty or very short (< 50 chars), set source_type to 'other', suggested_group to 'uncategorized', summary to empty string
- Keep summary factual and concise (2-3 sentences max)
- If the extracted title is clearly wrong (garbage text, error page title), suggest a better one`

	userPrompt := url + "\n\n" + content

	client := resty.New()
	reqBody := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.1,
		"max_tokens":  500,
	}

	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+apiKey).
		SetHeader("Content-Type", "application/json").
		SetBody(reqBody).
		SetResult(&map[string]interface{}{}).
		Post("https://api.openai.com/v1/chat/completions")

	if err != nil {
		return fallbackCategorize(url, content), nil
	}

	if resp.IsError() {
		return fallbackCategorize(url, content), nil
	}

	resultMap, ok := resp.Result().(*map[string]interface{})
	if !ok || resultMap == nil {
		return fallbackCategorize(url, content), nil
	}

	// Extract token usage — present in non-streaming OpenAI responses
	var inputTokens, outputTokens int
	if usage, ok := (*resultMap)["usage"].(map[string]interface{}); ok {
		if pt, ok := usage["prompt_tokens"].(float64); ok {
			inputTokens = int(pt)
		}
		if ct, ok := usage["completion_tokens"].(float64); ok {
			outputTokens = int(ct)
		}
	}

	// Extract content from response
	choices, ok := (*resultMap)["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return fallbackCategorize(url, content), nil
	}
	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return fallbackCategorize(url, content), nil
	}
	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return fallbackCategorize(url, content), nil
	}
	contentStr, ok := message["content"].(string)
	if !ok {
		return fallbackCategorize(url, content), nil
	}

	// Parse the JSON response from the LLM
	contentStr = strings.TrimSpace(contentStr)
	var result CategorizeResult
	if err := json.Unmarshal([]byte(contentStr), &result); err != nil {
		return fallbackCategorize(url, content), nil
	}

	// Validate fields
	if result.SourceType == "" {
		result.SourceType = "other"
	}
	// Guard against invented group names — snap to the allowed group's exact
	// casing (GetGroupByName matches exactly), or fall back to uncategorized.
	if g, ok := matchFold(groupsList, result.SuggestedGroup); ok {
		result.SuggestedGroup = g
	} else {
		result.SuggestedGroup = "uncategorized"
	}

	// Attach token usage and model
	result.InputTokens = inputTokens
	result.OutputTokens = outputTokens
	result.ModelUsed = model

	return &result, nil
}

// quoteJoin renders group names as a quoted, comma-separated JSON array fragment.
func quoteJoin(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = `"` + item + `"`
	}
	return strings.Join(quoted, ", ")
}

// matchFold returns the first item equal to s (case-insensitive) and whether it was found.
func matchFold(items []string, s string) (string, bool) {
	for _, item := range items {
		if strings.EqualFold(item, s) {
			return item, true
		}
	}
	return "", false
}

// fallbackCategorize provides a simple URL-based categorization without an LLM.
func fallbackCategorize(url, content string) *CategorizeResult {
	urlLower := strings.ToLower(url)
	result := &CategorizeResult{
		SourceType:     "other",
		SuggestedGroup: "uncategorized",
	}

	switch {
	case strings.Contains(urlLower, "github.com"):
		result.SourceType = "github"
		result.SuggestedGroup = "github"
	case strings.Contains(urlLower, "youtube.com") || strings.Contains(urlLower, "youtu.be") || strings.Contains(urlLower, "vimeo.com"):
		result.SourceType = "youtube"
		result.SuggestedGroup = "videos"
	case strings.Contains(urlLower, "twitter.com") || strings.Contains(urlLower, "x.com"):
		result.SourceType = "twitter"
		result.SuggestedGroup = "articles"
	case strings.Contains(urlLower, "docs.") || strings.Contains(urlLower, "documentation"):
		result.SourceType = "docs"
		result.SuggestedGroup = "docs"
	case strings.Contains(urlLower, "medium.com") || strings.Contains(urlLower, "blog.") || strings.Contains(urlLower, "/blog/"):
		result.SourceType = "article"
		result.SuggestedGroup = "articles"
	case strings.Contains(urlLower, "news.") || strings.Contains(urlLower, "/news/"):
		result.SourceType = "article"
		result.SuggestedGroup = "articles"
	}

	// Short content fallback
	if len(strings.TrimSpace(content)) < 50 {
		result.SourceType = "other"
		result.SuggestedGroup = "uncategorized"
		result.Summary = ""
	}

	return result
}
