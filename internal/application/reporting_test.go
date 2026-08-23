package application_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/audit"
	"github.com/wyw14/cry-081/internal/domain/identity"
)

func TestAuditExportAppliesPermissionAndResourceFilter(t *testing.T) {
	fixture := newWorkflowFixture(t)
	eventA, _ := audit.NewEvent("event-a", "editor1", "api", "schedule", "manuscript", "m1", "scheduled", map[string]string{"status": "accepted"}, map[string]string{"status": "scheduled"}, fixture.now)
	eventB, _ := audit.NewEvent("event-b", "editor1", "api", "withdraw", "article", "article-1", "correction", map[string]string{"status": "published"}, map[string]string{"status": "withdrawn"}, fixture.now)
	_ = fixture.audits.Append(context.Background(), eventA)
	_ = fixture.audits.Append(context.Background(), eventB)
	service := application.NewReportingService(fixture.manuscripts, fixture.reviews, fixture.audits)
	author, _ := fixture.users.Get(context.Background(), "author1")
	if err := service.ExportAuditCSV(context.Background(), author, application.AuditFilter{}, &bytes.Buffer{}); err == nil {
		t.Fatal("author exported audit events")
	}
	chief, _ := fixture.users.Get(context.Background(), "editor1")
	var output bytes.Buffer
	if err := service.ExportAuditCSV(context.Background(), chief, application.AuditFilter{Resource: "article", Limit: 10}, &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "event-b") || strings.Contains(output.String(), "event-a") {
		t.Fatalf("audit export ignored filter: %s", output.String())
	}
	if !chief.Can(identity.PermissionAuditRead) {
		t.Fatal("test fixture is missing audit permission")
	}
}
