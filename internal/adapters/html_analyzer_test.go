package adapters

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/architeacher/svc-web-analyzer/internal/domain"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTMLParser(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	parser := NewHTMLAnalyzer(&logger)

	assert.NotNil(t, parser)
	assert.Equal(t, &logger, parser.logger)
}

func TestHTMLParser_ExtractHTMLVersion(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	parser := NewHTMLAnalyzer(&logger)

	tests := []struct {
		name     string
		html     string
		expected domain.HTMLVersion
	}{
		{
			name:     "HTML5 doctype",
			html:     "<!DOCTYPE html><html><head><title>Test</title></head></html>",
			expected: domain.HTML5,
		},
		{
			name:     "HTML5 doctype with whitespace",
			html:     "<!DOCTYPE   html  ><html><head><title>Test</title></head></html>",
			expected: domain.HTML5,
		},
		{
			name:     "HTML5 doctype case insensitive",
			html:     "<!doctype HTML><html><head><title>Test</title></head></html>",
			expected: domain.HTML5,
		},
		{
			name:     "HTML 4.01 Strict",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01//EN"><html><head><title>Test</title></head></html>`,
			expected: domain.HTML401,
		},
		{
			name:     "HTML 4.01 Transitional",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"><html><head><title>Test</title></head></html>`,
			expected: domain.HTML401,
		},
		{
			name:     "XHTML 1.0 Strict",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN"><html><head><title>Test</title></head></html>`,
			expected: domain.XHTML10,
		},
		{
			name:     "XHTML 1.1",
			html:     `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN"><html><head><title>Test</title></head></html>`,
			expected: domain.XHTML11,
		},
		{
			name:     "XML declaration without doctype",
			html:     `<?xml version="1.0" encoding="UTF-8"?><html><head><title>Test</title></head></html>`,
			expected: domain.XHTML10,
		},
		{
			name:     "No doctype",
			html:     "<html><head><title>Test</title></head></html>",
			expected: domain.Unknown,
		},
		{
			name:     "Empty HTML",
			html:     "",
			expected: domain.Unknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parser.ExtractHTMLVersion(tt.html)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTMLParser_ExtractTitle(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	parser := NewHTMLParser(&logger)

	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Simple title",
			html:     "<html><head><title>Test Page</title></head></html>",
			expected: "Test Page",
		},
		{
			name:     "Title with extra whitespace",
			html:     "<html><head><title>  Test   Page  </title></head></html>",
			expected: "Test Page",
		},
		{
			name:     "Title with newlines",
			html:     "<html><head><title>\n  Test\n  Page\n  </title></head></html>",
			expected: "Test Page",
		},
		{
			name:     "Empty title",
			html:     "<html><head><title></title></head></html>",
			expected: "",
		},
		{
			name:     "No title tag",
			html:     "<html><head></head></html>",
			expected: "",
		},
		{
			name:     "Multiple title tags (should get first)",
			html:     "<html><head><title>First Title</title><title>Second Title</title></head></html>",
			expected: "First Title",
		},
		{
			name:     "Malformed HTML",
			html:     "<html><head><title>Test",
			expected: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parser.ExtractTitle(tt.html)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTMLParser_ExtractHeadingCounts(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	parser := NewHTMLParser(&logger)

	tests := []struct {
		name     string
		html     string
		expected domain.HeadingCounts
	}{
		{
			name: "All heading types",
			html: `<html><body>
				<h1>Heading 1</h1>
				<h1>Another H1</h1>
				<h2>Heading 2</h2>
				<h3>Heading 3</h3>
				<h3>Another H3</h3>
				<h3>Third H3</h3>
				<h4>Heading 4</h4>
				<h5>Heading 5</h5>
				<h6>Heading 6</h6>
			</body></html>`,
			expected: domain.HeadingCounts{
				H1: 2,
				H2: 1,
				H3: 3,
				H4: 1,
				H5: 1,
				H6: 1,
			},
		},
		{
			name: "No headings",
			html: "<html><body><p>No headings here</p></body></html>",
			expected: domain.HeadingCounts{
				H1: 0,
				H2: 0,
				H3: 0,
				H4: 0,
				H5: 0,
				H6: 0,
			},
		},
		{
			name: "Only H1 headings",
			html: `<html><body>
				<h1>First</h1>
				<h1>Second</h1>
				<h1>Third</h1>
			</body></html>`,
			expected: domain.HeadingCounts{
				H1: 3,
				H2: 0,
				H3: 0,
				H4: 0,
				H5: 0,
				H6: 0,
			},
		},
		{
			name:     "Malformed HTML",
			html:     "<h1>Test<h2>Another",
			expected: domain.HeadingCounts{H1: 1, H2: 1, H3: 0, H4: 0, H5: 0, H6: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parser.ExtractHeadingCounts(tt.html)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHTMLParser_ExtractLinks(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	parser := NewHTMLParser(&logger)

	tests := []struct {
		name     string
		html     string
		baseURL  string
		expected []domain.Link
		wantErr  bool
	}{
		{
			name: "Internal and external links",
			html: `<html><body>
				<a href="https://example.com/page1">Internal Link</a>
				<a href="https://other.com/page">External Link</a>
				<a href="/relative">Relative Link</a>
				<a href="relative2">Another Relative</a>
			</body></html>`,
			baseURL: "https://example.com",
			expected: []domain.Link{
				{URL: "https://example.com/page1", Type: domain.LinkTypeInternal},
				{URL: "https://other.com/page", Type: domain.LinkTypeExternal},
				{URL: "https://example.com/relative", Type: domain.LinkTypeInternal},
				{URL: "https://example.com/relative2", Type: domain.LinkTypeInternal},
			},
			wantErr: false,
		},
		{
			name: "Skip invalid links",
			html: `<html><body>
				<a href="https://example.com/valid">Valid Link</a>
				<a href="#fragment">Fragment</a>
				<a href="javascript:void(0)">JavaScript</a>
				<a href="mailto:test@example.com">Email</a>
				<a href="tel:+1234567890">Phone</a>
				<a href="">Empty</a>
				<a>No href</a>
			</body></html>`,
			baseURL: "https://example.com",
			expected: []domain.Link{
				{URL: "https://example.com/valid", Type: domain.LinkTypeInternal},
			},
			wantErr: false,
		},
		{
			name: "Duplicate links",
			html: `<html><body>
				<a href="https://example.com/page">Link 1</a>
				<a href="https://example.com/page">Link 2 (duplicate)</a>
				<a href="https://example.com/other">Other Link</a>
			</body></html>`,
			baseURL: "https://example.com",
			expected: []domain.Link{
				{URL: "https://example.com/page", Type: domain.LinkTypeInternal},
				{URL: "https://example.com/other", Type: domain.LinkTypeInternal},
			},
			wantErr: false,
		},
		{
			name:     "Invalid base URL",
			html:     `<a href="/test">Test</a>`,
			baseURL:  "://invalid",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "No links",
			html:     "<html><body><p>No links here</p></body></html>",
			baseURL:  "https://example.com",
			expected: []domain.Link{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := parser.ExtractLinks(tt.html, tt.baseURL)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, len(tt.expected), len(result))

			// Convert to maps for easier comparison (order might vary)
			expectedMap := make(map[string]domain.LinkType)
			for _, link := range tt.expected {
				expectedMap[link.URL] = link.Type
			}

			resultMap := make(map[string]domain.LinkType)
			for _, link := range result {
				resultMap[link.URL] = link.Type
			}

			assert.Equal(t, expectedMap, resultMap)
		})
	}
}

func TestHTMLParser_ExtractForms(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	parser := NewHTMLParser(&logger)

	tests := []struct {
		name     string
		html     string
		baseURL  string
		expected domain.FormAnalysis
	}{
		{
			name: "Login form with POST and password field",
			html: `<html><body>
				<form method="post" action="/login">
					<input type="text" name="username">
					<input type="password" name="password">
					<input type="submit" value="Login">
				</form>
			</body></html>`,
			baseURL: "https://example.com",
			expected: domain.FormAnalysis{
				TotalCount:         1,
				LoginFormsDetected: 1,
				LoginFormDetails: []domain.LoginForm{
					{
						Method: domain.FormMethod("POST"),
						Action: "https://example.com/login",
						Fields: []string{"username", "password"},
					},
				},
			},
		},
		{
			name: "Non-login form (GET method)",
			html: `<html><body>
				<form method="get" action="/search">
					<input type="text" name="query">
					<input type="submit" value="Search">
				</form>
			</body></html>`,
			baseURL: "https://example.com",
			expected: domain.FormAnalysis{
				TotalCount:         1,
				LoginFormsDetected: 0,
				LoginFormDetails:   []domain.LoginForm{},
			},
		},
		{
			name: "POST form without password field",
			html: `<html><body>
				<form method="post" action="/contact">
					<input type="text" name="name">
					<input type="email" name="email">
					<textarea name="message"></textarea>
					<input type="submit" value="Send">
				</form>
			</body></html>`,
			baseURL: "https://example.com",
			expected: domain.FormAnalysis{
				TotalCount:         1,
				LoginFormsDetected: 0,
				LoginFormDetails:   []domain.LoginForm{},
			},
		},
		{
			name: "Multiple forms with one login form",
			html: `<html><body>
				<form method="get" action="/search">
					<input type="text" name="query">
				</form>
				<form method="post" action="/auth">
					<input type="email" name="email">
					<input type="password" name="password">
				</form>
				<form method="post" action="/newsletter">
					<input type="email" name="email">
				</form>
			</body></html>`,
			baseURL: "https://example.com",
			expected: domain.FormAnalysis{
				TotalCount:         3,
				LoginFormsDetected: 1,
				LoginFormDetails: []domain.LoginForm{
					{
						Method: domain.FormMethod("POST"),
						Action: "https://example.com/auth",
						Fields: []string{"email", "password"},
					},
				},
			},
		},
		{
			name: "Form with default method (GET)",
			html: `<html><body>
				<form action="/search">
					<input type="text" name="query">
				</form>
			</body></html>`,
			baseURL: "https://example.com",
			expected: domain.FormAnalysis{
				TotalCount:         1,
				LoginFormsDetected: 0,
				LoginFormDetails:   []domain.LoginForm{},
			},
		},
		{
			name: "Form with relative action URL",
			html: `<html><body>
				<form method="post" action="login">
					<input type="password" name="pass">
				</form>
			</body></html>`,
			baseURL: "https://example.com/app/",
			expected: domain.FormAnalysis{
				TotalCount:         1,
				LoginFormsDetected: 1,
				LoginFormDetails: []domain.LoginForm{
					{
						Method: domain.FormMethod("POST"),
						Action: "https://example.com/app/login",
						Fields: []string{"pass"},
					},
				},
			},
		},
		{
			name:    "No forms",
			html:    "<html><body><p>No forms here</p></body></html>",
			baseURL: "https://example.com",
			expected: domain.FormAnalysis{
				TotalCount:         0,
				LoginFormsDetected: 0,
				LoginFormDetails:   []domain.LoginForm{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parser.ExtractForms(tt.html, tt.baseURL)
			assert.Equal(t, tt.expected.TotalCount, result.TotalCount)
			assert.Equal(t, tt.expected.LoginFormsDetected, result.LoginFormsDetected)
			assert.Equal(t, len(tt.expected.LoginFormDetails), len(result.LoginFormDetails))

			// Check each login form detail
			for i, expected := range tt.expected.LoginFormDetails {
				if i < len(result.LoginFormDetails) {
					actual := result.LoginFormDetails[i]
					assert.Equal(t, expected.Method, actual.Method)
					assert.Equal(t, expected.Action, actual.Action)
					assert.ElementsMatch(t, expected.Fields, actual.Fields)
				}
			}
		})
	}
}

func TestHTMLParser_isLikelyLoginForm(t *testing.T) {
	t.Parallel()

	logger := zerolog.Nop()
	parser := NewHTMLParser(&logger)

	tests := []struct {
		name     string
		method   string
		html     string
		expected bool
	}{
		{
			name:     "POST form with password field",
			method:   "POST",
			html:     `<form><input type="password" name="password"></form>`,
			expected: true,
		},
		{
			name:     "GET form with password field",
			method:   "GET",
			html:     `<form><input type="password" name="password"></form>`,
			expected: false,
		},
		{
			name:     "POST form without password field",
			method:   "POST",
			html:     `<form><input type="text" name="username"></form>`,
			expected: false,
		},
		{
			name:     "POST form with multiple password fields",
			method:   "POST",
			html:     `<form><input type="password" name="password"><input type="password" name="confirm"></form>`,
			expected: true,
		},
		{
			name:     "Empty form",
			method:   "POST",
			html:     `<form></form>`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We need to create a goquery selection for the test
			// This is a bit hacky but necessary for testing the internal method
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tt.html))
			require.NoError(t, err)

			selection := doc.Find("form").First()
			t.Parallel()
			result := parser.isLikelyLoginForm(tt.method, selection)
			assert.Equal(t, tt.expected, result)
		})
	}
}
