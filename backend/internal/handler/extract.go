package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type ExtractHandler struct{}

func NewExtractHandler() *ExtractHandler { return &ExtractHandler{} }

type extractRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

type extractResponse struct {
	Company  *string `json:"company"`
	Role     *string `json:"role"`
	Location *string `json:"location"`
	Source   string  `json:"source"`
}

func (h *ExtractHandler) Extract(w http.ResponseWriter, r *http.Request) {
	var req extractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "url is required")
		return
	}

	if result := tryQuickParse(req.Title, req.URL); result != nil {
		result.Location = extractLocation(req.Content)
		writeJSON(w, http.StatusOK, result)
		return
	}

	result, err := callGemini(req.Title, req.Content, req.URL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("extraction failed: %s", err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func looksLikeCompany(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 80 {
		return false
	}
	if regexp.MustCompile(`^[\$\x{20AC}\x{00A3}\x{00A5}\x{20A6}\d]`).MatchString(s) {
		return false
	}
	if strings.HasSuffix(s, ")") && !strings.Contains(s, "(") {
		return false
	}
	lower := strings.ToLower(s)
	badWords := []string{
		"remote", "hybrid", "onsite", "on-site", "entry", "senior", "junior",
		"mid-level", "lead", "staff", "principal", "contract", "full-time",
		"part-time", "freelance", "temporary", "intern",
	}
	for _, bad := range badWords {
		if strings.HasPrefix(lower, bad) {
			return false
		}
	}
	return true
}

func tryQuickParse(title, url string) *extractResponse {
	if title == "" {
		return nil
	}

	atRe := regexp.MustCompile(`(?i)^(.+?)\s+at\s+(.+?)(?:\s*[|` + "\u2013" + `\-].*)?$`)
	if m := atRe.FindStringSubmatch(title); len(m) == 3 {
		role := strings.TrimSpace(m[1])
		company := strings.TrimSpace(m[2])
		if looksLikeCompany(company) && len(role) > 2 {
			return &extractResponse{Company: &company, Role: &role, Source: detectSource(url)}
		}
	}

	sepRe := regexp.MustCompile(`^(.+?)\s*[-` + "\u2013" + `]\s*(.+?)(?:\s*[|` + "\u2013" + `\-].*)?$`)
	if m := sepRe.FindStringSubmatch(title); len(m) == 3 {
		role := strings.TrimSpace(m[1])
		company := regexp.MustCompile(`(?i)\s*(careers?|jobs?|hiring)\s*`).
			ReplaceAllString(strings.TrimSpace(m[2]), "")
		company = strings.TrimSpace(company)
		if looksLikeCompany(company) && len(role) > 2 {
			return &extractResponse{Company: &company, Role: &role, Source: detectSource(url)}
		}
	}

	return nil
}

func extractLocation(content string) *string {
	if len(content) > 1500 {
		content = content[:1500]
	}
	re := regexp.MustCompile(`\b(Remote|Hybrid|On.?site|[A-Z][a-z]{2,}(?:,\s*[A-Z][a-z]{2,})?)\b`)
	if m := re.FindString(content); m != "" {
		return &m
	}
	return nil
}

func detectSource(url string) string {
	switch {
	case strings.Contains(url, "linkedin"):
		return "linkedin"
	case strings.Contains(url, "greenhouse"):
		return "job_board"
	case strings.Contains(url, "lever"):
		return "job_board"
	case strings.Contains(url, "indeed"):
		return "job_board"
	case strings.Contains(url, "glassdoor"):
		return "job_board"
	case strings.Contains(url, "wellfound"):
		return "job_board"
	case strings.Contains(url, "workday"):
		return "job_board"
	case strings.Contains(url, "workable"):
		return "job_board"
	case strings.Contains(url, "crossover"):
		return "job_board"
	case strings.Contains(url, "smartrecruiters"):
		return "job_board"
	case strings.Contains(url, "ashbyhq"):
		return "job_board"
	default:
		return "company_site"
	}
}

func callGemini(title, content, url string) (*extractResponse, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	if len(content) > 6000 {
		content = content[:6000]
	}

	prompt := "You are extracting structured data from a job posting page. Return ONLY valid JSON, no explanation, no markdown.\n\n" +
		"Page URL: " + url + "\n" +
		"Page title: " + title + "\n\n" +
		"Page content:\n" + content + "\n\n" +
		`Return exactly this JSON:
{
  "company": "legal name of the hiring employer",
  "role": "exact job title only",
  "location": "city/country or Remote or null",
  "source": "linkedin|job_board|company_site|recruiter|referral|other"
}

STRICT RULES:
- company: NEVER put salary, location, job level, or job board names (Greenhouse, Lever, Crossover, Workable, Canonical-but-on-greenhouse). Look in the page CONTENT for the actual employer.
- role: Job title only. No salary, no company, no level suffix like "(Entry-Level)".
- location: City/country or "Remote". Never a company name.
- source: "job_board" if URL has greenhouse/lever/workable/crossover/indeed/glassdoor. Otherwise "company_site".

Examples:
- Canonical page on greenhouse → company:"Canonical", role:"Ubuntu Sales Engineer", location:"Remote"
- Moniepoint careers page → company:"Moniepoint", role:"Site Reliability Engineer", location:"Remote, Nigeria"
- Crossover DevOps → company:"IgniteTech", role:"DevOps Specialist", location:"Remote"
- Servant Talent on workable → company:"Servant Talent", role:"DevOps & Cloud Engineer"}`

	reqBody := map[string]any{
		"contents": []map[string]any{
			{"parts": []map[string]any{{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"temperature":      0,
			"maxOutputTokens":  300,
			"responseMimeType": "application/json",
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	geminiURL := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/gemini-3.1-flash-lite-preview:generateContent?key=%s",
		apiKey,
	)

	resp, err := http.Post(geminiURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("parsing gemini response: %w", err)
	}
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini: %s", geminiResp.Error.Message)
	}
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini returned no candidates")
	}

	raw := strings.TrimSpace(geminiResp.Candidates[0].Content.Parts[0].Text)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var result extractResponse
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, fmt.Errorf("parsing extracted JSON: %w", err)
	}
	if result.Source == "" {
		result.Source = detectSource(url)
	}

	return &result, nil
}
