package manuscript

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"time"

	"github.com/wyw14/cry-081/internal/domain/shared"
)

type Attachment struct {
	ID       string
	Name     string
	MIME     string
	Size     int64
	Digest   string
	StoreKey string
}

func (a Attachment) Valid() bool {
	return a.ID != "" && a.Name != "" && a.MIME != "" && a.Size > 0 && a.Digest != "" && a.StoreKey != ""
}

type DraftVersion struct {
	Number      int64
	Title       string
	Abstract    string
	Markdown    string
	Tags        []string
	Attachments []Attachment
	Fingerprint string
	CreatedAt   time.Time
	Locked      bool
	Correction  bool
}

func NewDraftVersion(number int64, title, abstract, markdown string, tags []string, attachments []Attachment, now time.Time) (DraftVersion, error) {
	if number < 1 || strings.TrimSpace(title) == "" || strings.TrimSpace(abstract) == "" || strings.TrimSpace(markdown) == "" {
		return DraftVersion{}, shared.NewError("DRAFT_INVALID", "draft content is incomplete", shared.ErrValidation)
	}
	if len(markdown) < 80 {
		return DraftVersion{}, shared.NewError("DRAFT_TOO_SHORT", "markdown body is too short", shared.ErrValidation)
	}
	cleanTags := normalizeTags(tags)
	cleanAttachments := make([]Attachment, len(attachments))
	copy(cleanAttachments, attachments)
	for _, attachment := range cleanAttachments {
		if !attachment.Valid() {
			return DraftVersion{}, shared.NewError("ATTACHMENT_INVALID", "attachment metadata is incomplete", shared.ErrValidation)
		}
	}
	version := DraftVersion{
		Number: number, Title: strings.TrimSpace(title), Abstract: strings.TrimSpace(abstract),
		Markdown: markdown, Tags: cleanTags, Attachments: cleanAttachments, CreatedAt: now.UTC(),
	}
	version.Fingerprint = version.ContentDigest()
	return version, nil
}

func normalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	clean := make([]string, 0, len(tags))
	for _, tag := range tags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		clean = append(clean, tag)
	}
	sort.Strings(clean)
	return clean
}

func (v DraftVersion) ContentDigest() string {
	h := sha256.New()
	h.Write([]byte(v.Title))
	h.Write([]byte{0})
	h.Write([]byte(v.Abstract))
	h.Write([]byte{0})
	h.Write([]byte(v.Markdown))
	for _, tag := range v.Tags {
		h.Write([]byte{0})
		h.Write([]byte(tag))
	}
	for _, attachment := range v.Attachments {
		h.Write([]byte{0})
		h.Write([]byte(attachment.Digest))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func (v DraftVersion) Clone() DraftVersion {
	v.Tags = append([]string(nil), v.Tags...)
	v.Attachments = append([]Attachment(nil), v.Attachments...)
	return v
}
