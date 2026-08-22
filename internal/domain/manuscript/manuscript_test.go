package manuscript_test

import (
	"testing"
	"time"

	"github.com/wyw14/cry-081/internal/domain/manuscript"
)

func TestSubmissionSnapshotRemainsImmutableAcrossDraftChanges(t *testing.T) {
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	version, err := manuscript.NewDraftVersion(1, "可复现研究", "一段完整摘要", longBody(), []string{"Research", "Go"}, []manuscript.Attachment{{ID: "a1", Name: "data.pdf", MIME: "application/pdf", Size: 256, Digest: "digest-a", StoreKey: "m1/a1"}}, now)
	if err != nil {
		t.Fatal(err)
	}
	target, err := manuscript.New("m1", "author1", "research", version, now)
	if err != nil {
		t.Fatal(err)
	}
	snapshot, err := manuscript.NewSubmissionSnapshot("s1", target.ID, target.AuthorID, version, manuscript.Declaration{OriginalWork: true, AuthorApproved: true, EthicsCleared: true, Text: "all authors agree"}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := target.Submit(snapshot, now); err != nil {
		t.Fatal(err)
	}
	version.Tags[0] = "changed"
	version.Attachments[0].Name = "changed.pdf"
	if target.Submissions[0].Tags[0] != "go" || target.Submissions[0].Attachments[0].Name != "data.pdf" {
		t.Fatalf("submission snapshot changed with source draft: %#v", target.Submissions[0])
	}
	if !target.Versions[0].Locked {
		t.Fatal("submitted version must be locked")
	}
}

func TestAcceptedManuscriptOnlyAllowsMarkedCorrection(t *testing.T) {
	now := time.Date(2026, 8, 23, 8, 0, 0, 0, time.UTC)
	version, _ := manuscript.NewDraftVersion(1, "稿件标题", "完整摘要", longBody(), nil, nil, now)
	target, _ := manuscript.New("m1", "author1", "research", version, now)
	for _, status := range []manuscript.Status{manuscript.StatusSubmitted, manuscript.StatusUnderInitial, manuscript.StatusUnderFinal, manuscript.StatusAccepted} {
		if err := target.Transition(status, "editor1", "workflow progression", target.ActiveVersion, now); err != nil {
			t.Fatal(err)
		}
	}
	next, _ := manuscript.NewDraftVersion(2, "勘误标题", "勘误摘要", longBody(), nil, nil, now.Add(time.Hour))
	if err := target.AddDraft(next, target.AuthorID, now); err == nil {
		t.Fatal("accepted manuscript accepted an ordinary draft")
	}
	next.Correction = true
	if err := target.AddCorrection(next, target.AuthorID, now); err != nil {
		t.Fatalf("marked correction should be accepted: %v", err)
	}
}

func TestStateMachineRejectsIllegalJump(t *testing.T) {
	now := time.Now().UTC()
	version, _ := manuscript.NewDraftVersion(1, "稿件标题", "完整摘要", longBody(), nil, nil, now)
	target, _ := manuscript.New("m1", "author1", "research", version, now)
	if err := target.Transition(manuscript.StatusAccepted, "chief1", "looks good", 1, now); err == nil {
		t.Fatal("draft jumped directly to accepted")
	}
}

func longBody() string {
	return "研究正文包含足够的结构化内容，用于说明问题背景、研究方法、样本选择、实验过程、结果分析、局限讨论和结论建议。This manuscript body contains enough words to satisfy the deterministic format preflight policy for domain behavior tests and submission snapshots."
}
