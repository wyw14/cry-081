package identity

import "github.com/wyw14/cry-081/internal/domain/shared"

type Role string

const (
	RoleAuthor         Role = "author"
	RoleEditor         Role = "editor"
	RoleChiefEditor    Role = "chief_editor"
	RoleSectionManager Role = "section_manager"
)

var validRoles = map[Role]struct{}{
	RoleAuthor: {}, RoleEditor: {}, RoleChiefEditor: {}, RoleSectionManager: {},
}

func (r Role) Valid() bool {
	_, ok := validRoles[r]
	return ok
}

func ParseRole(raw string) (Role, error) {
	role := Role(raw)
	if !role.Valid() {
		return "", shared.NewError("ROLE_INVALID", "role is not supported", shared.ErrValidation)
	}
	return role, nil
}

type Permission string

const (
	PermissionDraftWrite       Permission = "draft:write"
	PermissionSubmissionRead   Permission = "submission:read"
	PermissionAssignmentManage Permission = "assignment:manage"
	PermissionReviewWrite      Permission = "review:write"
	PermissionDecisionWrite    Permission = "decision:write"
	PermissionIssueManage      Permission = "issue:manage"
	PermissionPublicationWrite Permission = "publication:write"
	PermissionReportRead       Permission = "report:read"
	PermissionAuditRead        Permission = "audit:read"
)

var permissionsByRole = map[Role]map[Permission]struct{}{
	RoleAuthor: {
		PermissionDraftWrite: {}, PermissionSubmissionRead: {},
	},
	RoleEditor: {
		PermissionSubmissionRead: {}, PermissionReviewWrite: {}, PermissionReportRead: {},
	},
	RoleChiefEditor: {
		PermissionSubmissionRead: {}, PermissionAssignmentManage: {}, PermissionReviewWrite: {},
		PermissionDecisionWrite: {}, PermissionIssueManage: {}, PermissionPublicationWrite: {},
		PermissionReportRead: {}, PermissionAuditRead: {},
	},
	RoleSectionManager: {
		PermissionSubmissionRead: {}, PermissionAssignmentManage: {}, PermissionIssueManage: {},
		PermissionPublicationWrite: {}, PermissionReportRead: {}, PermissionAuditRead: {},
	},
}

func Allows(role Role, permission Permission) bool {
	permissions, ok := permissionsByRole[role]
	if !ok {
		return false
	}
	_, ok = permissions[permission]
	return ok
}
