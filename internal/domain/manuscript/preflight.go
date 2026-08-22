package manuscript

import (
	"fmt"
	"strings"
)

type PreflightResult struct {
	Passed     bool
	WordCount  int
	Violations []string
}

type FormatPolicy struct {
	MinimumWords       int
	MaximumAttachments int
	AllowedMIMEs       map[string]struct{}
	MaximumFileSize    int64
}

func DefaultFormatPolicy() FormatPolicy {
	return FormatPolicy{
		MinimumWords: 20, MaximumAttachments: 8, MaximumFileSize: 10 << 20,
		AllowedMIMEs: map[string]struct{}{
			"application/pdf": {}, "image/png": {}, "image/jpeg": {}, "text/plain": {},
		},
	}
}

func (p FormatPolicy) Check(version DraftVersion) PreflightResult {
	result := PreflightResult{Passed: true, WordCount: len(strings.Fields(version.Markdown))}
	if result.WordCount < p.MinimumWords {
		result.Violations = append(result.Violations, fmt.Sprintf("body requires at least %d words", p.MinimumWords))
	}
	if len(version.Attachments) > p.MaximumAttachments {
		result.Violations = append(result.Violations, "too many attachments")
	}
	for _, attachment := range version.Attachments {
		if _, ok := p.AllowedMIMEs[attachment.MIME]; !ok {
			result.Violations = append(result.Violations, "attachment MIME is not allowed")
		}
		if attachment.Size > p.MaximumFileSize {
			result.Violations = append(result.Violations, "attachment is too large")
		}
	}
	result.Passed = len(result.Violations) == 0
	return result
}
