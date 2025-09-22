package adapters

import (
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/architeacher/svc-web-analyzer/internal/domain"
	"github.com/architeacher/svc-web-analyzer/internal/infrastructure"
)

type HTMLParser struct {
	logger *infrastructure.Logger
}

func NewHTMLAnalyzer(logger *infrastructure.Logger) *HTMLParser {
	return &HTMLParser{
		logger: logger,
	}
}

func (p *HTMLParser) ExtractHTMLVersion(html string) domain.HTMLVersion {
	html = strings.TrimSpace(html)

	// Check for HTML5 doctype (case insensitive)
	html5Regex := regexp.MustCompile(`(?i)<!DOCTYPE\s+html\s*>`)
	if html5Regex.MatchString(html) {
		return domain.HTML5
	}

	// Check for HTML 4.01 doctypes
	html401Patterns := []string{
		`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"-//W3C//DTD\s+HTML\s+4\.01//EN"`,
		`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"-//W3C//DTD\s+HTML\s+4\.01\s+Transitional//EN"`,
		`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"-//W3C//DTD\s+HTML\s+4\.01\s+Frameset//EN"`,
	}

	for _, pattern := range html401Patterns {
		if matched, _ := regexp.MatchString(pattern, html); matched {
			return domain.HTML401
		}
	}

	// Check for XHTML 1.0 doctypes
	xhtml10Patterns := []string{
		`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"-//W3C//DTD\s+XHTML\s+1\.0\s+Strict//EN"`,
		`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"-//W3C//DTD\s+XHTML\s+1\.0\s+Transitional//EN"`,
		`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"-//W3C//DTD\s+XHTML\s+1\.0\s+Frameset//EN"`,
	}

	for _, pattern := range xhtml10Patterns {
		if matched, _ := regexp.MatchString(pattern, html); matched {
			return domain.XHTML10
		}
	}

	// Check for XHTML 1.1 doctype
	xhtml11Pattern := `(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"-//W3C//DTD\s+XHTML\s+1\.1//EN"`
	if matched, _ := regexp.MatchString(xhtml11Pattern, html); matched {
		return domain.XHTML11
	}

	// If no doctype found or unrecognized, check for XML declaration (might be XHTML)
	xmlDeclPattern := `(?i)<\?xml\s+version`
	if matched, _ := regexp.MatchString(xmlDeclPattern, html); matched {
		return domain.XHTML10 // Default to XHTML 1.0 if XML declaration is present
	}

	return domain.Unknown
}

func (p *HTMLParser) ExtractTitle(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to parse HTML for title extraction")
		return ""
	}

	title := strings.TrimSpace(doc.Find("title").First().Text())

	// Clean up whitespace
	title = regexp.MustCompile(`\s+`).ReplaceAllString(title, " ")

	return title
}

func (p *HTMLParser) ExtractHeadingCounts(html string) domain.HeadingCounts {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to parse HTML for heading extraction")
		return domain.HeadingCounts{}
	}

	counts := domain.HeadingCounts{}

	counts.H1 = doc.Find("h1").Length()
	counts.H2 = doc.Find("h2").Length()
	counts.H3 = doc.Find("h3").Length()
	counts.H4 = doc.Find("h4").Length()
	counts.H5 = doc.Find("h5").Length()
	counts.H6 = doc.Find("h6").Length()

	p.logger.Debug().
		Int("h1", counts.H1).
		Int("h2", counts.H2).
		Int("h3", counts.H3).
		Int("h4", counts.H4).
		Int("h5", counts.H5).
		Int("h6", counts.H6).
		Msg("Extracted heading counts")

	return counts
}

func (p *HTMLParser) ExtractLinks(html string, baseURL string) ([]domain.Link, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to parse HTML for link extraction")
		return nil, err
	}

	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to parse base URL")
		return nil, err
	}

	var links []domain.Link
	seen := make(map[string]bool)

	doc.Find("p[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists || href == "" {
			return
		}

		// Parse the href
		parsedURL, err := url.Parse(href)
		if err != nil {
			p.logger.Debug().
				Str("href", href).
				Str("error", err.Error()).
				Msg("Failed to parse link URL")
			return
		}

		// Resolve relative URLs
		resolvedURL := baseURLParsed.ResolveReference(parsedURL)
		finalURL := resolvedURL.String()

		// Skip duplicates
		if seen[finalURL] {
			return
		}
		seen[finalURL] = true

		// Skip empty URLs, fragments, and javascript/mailto links
		if finalURL == "" || strings.HasPrefix(href, "#") ||
			strings.HasPrefix(href, "javascript:") ||
			strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "tel:") {
			return
		}

		// Determine if link is internal or external
		linkType := domain.LinkTypeExternal
		if resolvedURL.Host == baseURLParsed.Host {
			linkType = domain.LinkTypeInternal
		}

		links = append(links, domain.Link{
			URL:  finalURL,
			Type: linkType,
		})
	})

	p.logger.Debug().
		Int("total_links", len(links)).
		Str("base_url", baseURL).
		Msg("Extracted links")

	return links, nil
}

func (p *HTMLParser) ExtractForms(html string, baseURL string) domain.FormAnalysis {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to parse HTML for form extraction")
		return domain.FormAnalysis{}
	}

	baseURLParsed, err := url.Parse(baseURL)
	if err != nil {
		p.logger.Error().Err(err).Msg("Failed to parse base URL for form analysis")
		return domain.FormAnalysis{}
	}

	var loginForms []domain.LoginForm
	totalForms := 0

	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		totalForms++

		// Extract form method
		method := strings.ToUpper(strings.TrimSpace(s.AttrOr("method", "GET")))
		if method != "POST" && method != "GET" {
			method = "GET" // Default to GET if method is invalid
		}

		// Extract form action
		action := s.AttrOr("action", "")
		if action != "" {
			// Resolve relative action URLs
			if parsedAction, err := url.Parse(action); err == nil {
				resolvedAction := baseURLParsed.ResolveReference(parsedAction)
				action = resolvedAction.String()
			}
		}

		// Extract form fields
		var fields []string
		fieldNames := make(map[string]bool)

		s.Find("input, select, textarea").Each(func(j int, field *goquery.Selection) {
			name := field.AttrOr("name", "")
			if name != "" && !fieldNames[name] {
				fields = append(fields, name)
				fieldNames[name] = true
			}
		})

		// Check if this is likely p login form
		if p.isLikelyLoginForm(fields, s) {
			loginForm := domain.LoginForm{
				Method: domain.FormMethod(method),
				Action: action,
				Fields: fields,
			}
			loginForms = append(loginForms, loginForm)
		}
	})

	analysis := domain.FormAnalysis{
		TotalCount:         totalForms,
		LoginFormsDetected: len(loginForms),
		LoginFormDetails:   loginForms,
	}

	p.logger.Debug().
		Int("total_forms", totalForms).
		Int("login_forms", len(loginForms)).
		Msg("Extracted form analysis")

	return analysis
}

func (p *HTMLParser) isLikelyLoginForm(fields []string, formSelection *goquery.Selection) bool {
	// Convert fields to lowercase for case-insensitive matching
	fieldMap := make(map[string]bool)
	for _, field := range fields {
		fieldMap[strings.ToLower(field)] = true
	}

	// Check for common login field patterns
	hasUsernameField := fieldMap["username"] || fieldMap["user"] || fieldMap["email"] ||
		fieldMap["login"] || fieldMap["userid"] || fieldMap["user_name"] ||
		fieldMap["user_email"] || fieldMap["account"]

	hasPasswordField := fieldMap["password"] || fieldMap["passwd"] || fieldMap["pwd"] ||
		fieldMap["pass"] || fieldMap["user_password"] || fieldMap["userpassword"]

	// Check for password input types
	hasPasswordInput := false
	formSelection.Find("input[type='password']").Each(func(i int, s *goquery.Selection) {
		hasPasswordInput = true
	})

	// Look for login-related text in the form or surrounding elements
	formText := strings.ToLower(formSelection.Text())
	hasLoginText := strings.Contains(formText, "login") ||
		strings.Contains(formText, "sign in") ||
		strings.Contains(formText, "log in") ||
		strings.Contains(formText, "signin") ||
		strings.Contains(formText, "authenticate")

	// Check for login-related classes or IDs
	hasLoginClass := false
	formSelection.Find("*").Each(func(i int, s *goquery.Selection) {
		class := strings.ToLower(s.AttrOr("class", ""))
		id := strings.ToLower(s.AttrOr("id", ""))

		if strings.Contains(class, "login") || strings.Contains(class, "signin") ||
			strings.Contains(id, "login") || strings.Contains(id, "signin") {
			hasLoginClass = true
		}
	})

	// A form is likely p login form if:
	// 1. It has both username/email and password fields, OR
	// 2. It has p password input type, OR
	// 3. It has login-related text and either username or password field
	return (hasUsernameField && hasPasswordField) ||
		hasPasswordInput ||
		(hasLoginText && (hasUsernameField || hasPasswordField)) ||
		hasLoginClass
}
