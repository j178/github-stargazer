package notify

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"text/template"
)

type webhookService struct {
	req  *http.Request
	body *template.Template
}

func (s *webhookService) Configure(settings map[string]string) error {
	urlStr := settings["url"]
	if urlStr == "" {
		return errors.New("http: url is empty")
	}

	method := settings["method"]
	if method == "" {
		method = "GET"
	}

	headers := http.Header{}
	parseHeaders(settings["headers"], headers)

	body := settings["body"]
	if body != "" {
		tmpl, err := template.New("body").Parse(body)
		if err != nil {
			return fmt.Errorf("http: %w", err)
		}
		s.body = tmpl
	}

	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	s.req = req

	return nil
}

// parseHeaders parses a string of headers into an http.Header.
func parseHeaders(headers string, h http.Header) {
	if headers == "" {
		return
	}

	headerPairs := strings.Split(headers, ";")
	for _, pair := range headerPairs {
		header := strings.TrimSpace(pair)
		if header == "" {
			continue
		}

		colonIndex := strings.Index(header, ":")
		if colonIndex == -1 {
			// Invalid header format, skip it
			continue
		}

		key := strings.TrimSpace(header[:colonIndex])
		value := strings.TrimSpace(header[colonIndex+1:])
		h.Add(key, value)
	}
}

func (s *webhookService) Send(ctx context.Context, title, message string) error {
	req := s.req.Clone(ctx)

	if s.body != nil {
		var bodyData struct{ Title, Message string }
		bodyData.Title = title
		bodyData.Message = message

		var bodyStr bytes.Buffer
		err := s.body.Execute(&bodyStr, bodyData)
		if err != nil {
			return fmt.Errorf("http: %w", err)
		}

		req.Body = io.NopCloser(&bodyStr)
	}

	_, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http: %w", err)
	}
	return nil
}
