package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AnalysisStatus string

const (
	StatusRequested  AnalysisStatus = "requested"
	StatusInProgress AnalysisStatus = "in_progress"
	StatusCompleted  AnalysisStatus = "completed"
	StatusFailed     AnalysisStatus = "failed"
)

type HTMLVersion string

const (
	HTML5   HTMLVersion = "HTML5"
	HTML401 HTMLVersion = "HTML 4.01"
	XHTML10 HTMLVersion = "XHTML 1.0"
	XHTML11 HTMLVersion = "XHTML 1.1"
	Unknown HTMLVersion = "Unknown"
)

type LinkType string

const (
	LinkTypeInternal LinkType = "internal"
	LinkTypeExternal LinkType = "external"
)

type FormMethod string

const (
	FormMethodPOST FormMethod = "POST"
	FormMethodGET  FormMethod = "GET"
)

type Analysis struct {
	ID          uuid.UUID      `json:"analysis_id"`
	URL         string         `json:"url"`
	Status      AnalysisStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
	Duration    *time.Duration `json:"duration,omitempty"`
	Results     *AnalysisData  `json:"results,omitempty"`
	Error       *AnalysisError `json:"error,omitempty"`
}

type AnalysisData struct {
	HTMLVersion   HTMLVersion   `json:"html_version"`
	Title         string        `json:"title"`
	HeadingCounts HeadingCounts `json:"heading_counts"`
	Links         LinkAnalysis  `json:"links"`
	Forms         FormAnalysis  `json:"forms"`
}

type HeadingCounts struct {
	H1 int `json:"h1"`
	H2 int `json:"h2"`
	H3 int `json:"h3"`
	H4 int `json:"h4"`
	H5 int `json:"h5"`
	H6 int `json:"h6"`
}

type LinkAnalysis struct {
	InternalCount     int                `json:"internal_count"`
	ExternalCount     int                `json:"external_count"`
	TotalCount        int                `json:"total_count"`
	InaccessibleLinks []InaccessibleLink `json:"inaccessible_links"`
}

type InaccessibleLink struct {
	URL        string `json:"url"`
	StatusCode int    `json:"status_code"`
	Error      string `json:"error"`
}

type FormAnalysis struct {
	TotalCount         int         `json:"total_count"`
	LoginFormsDetected int         `json:"login_forms_detected"`
	LoginFormDetails   []LoginForm `json:"login_form_details"`
}

type LoginForm struct {
	Method FormMethod `json:"method"`
	Action string     `json:"action"`
	Fields []string   `json:"fields"`
}

type AnalysisError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code,omitempty"`
	Details    string `json:"details,omitempty"`
}

type AnalysisOptions struct {
	IncludeHeadings bool          `json:"include_headings"`
	CheckLinks      bool          `json:"check_links"`
	DetectForms     bool          `json:"detect_forms"`
	Timeout         time.Duration `json:"timeout"`
}

type WebPageAnalyzer interface {
	Analyze(ctx context.Context, url string, options AnalysisOptions) (*AnalysisData, error)
}

type HTMLAnalyzer interface {
	ExtractHTMLVersion(html string) HTMLVersion
	ExtractTitle(html string) string
	ExtractHeadingCounts(html string) HeadingCounts
	ExtractLinks(html string, baseURL string) ([]Link, error)
	ExtractForms(html string, baseURL string) FormAnalysis
}

type Link struct {
	URL  string
	Type LinkType
}

type WebPageContent struct {
	URL         string
	StatusCode  int
	HTML        string
	ContentType string
	Headers     map[string]string
}

type LinkChecker interface {
	CheckAccessibility(ctx context.Context, links []Link) []InaccessibleLink
}

type CacheService interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, expiry time.Duration) error
	Delete(ctx context.Context, key string) error
}
type AnalysisEvent struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	EventID string      `json:"event_id"`
}

const (
	EventTypeStarted   = "analysis_started"
	EventTypeProgress  = "analysis_progress"
	EventTypeCompleted = "analysis_completed"
	EventTypeFailed    = "analysis_failed"
)
