package httptransport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/domain/identity"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/domain/review"
	"github.com/wyw14/cry-081/internal/domain/shared"
	"github.com/wyw14/cry-081/internal/middleware"
)

type Handlers struct {
	identity     *application.IdentityService
	manuscripts  *application.ManuscriptService
	reviews      *application.ReviewService
	publications *application.PublicationService
	reporting    *application.ReportingService
	validate     *validator.Validate
}

func NewHandlers(identityService *application.IdentityService, manuscriptService *application.ManuscriptService, reviewService *application.ReviewService, publicationService *application.PublicationService, reportingService *application.ReportingService) *Handlers {
	return &Handlers{identity: identityService, manuscripts: manuscriptService, reviews: reviewService, publications: publicationService, reporting: reportingService, validate: validator.New()}
}

type registerRequest struct {
	Email       string          `json:"email" validate:"required,email"`
	DisplayName string          `json:"display_name" validate:"required,min=2,max=80"`
	Password    string          `json:"password" validate:"required,min=10,max=128"`
	Roles       []identity.Role `json:"roles" validate:"required,min=1"`
}

func (h *Handlers) Register(c *gin.Context) {
	request, valid := decodeRequest[registerRequest](h, c, "request body is invalid")
	if !valid {
		return
	}
	user, err := h.identity.Register(c.Request.Context(), application.RegisterUserInput{Email: request.Email, DisplayName: request.DisplayName, Password: request.Password, Roles: request.Roles})
	respond(c, http.StatusCreated, gin.H{"id": user.ID, "email": user.Email, "display_name": user.DisplayName, "roles": user.Roles}, err)
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *Handlers) Login(c *gin.Context) {
	request, valid := decodeRequest[loginRequest](h, c, "email and password are required")
	if !valid {
		return
	}
	session, err := h.identity.Login(c.Request.Context(), request.Email, request.Password)
	respond(c, http.StatusOK, gin.H{"access_token": session.AccessToken, "refresh_token": session.RefreshToken, "refresh_expires_at": session.ExpiresAt}, err)
}

type createManuscriptRequest struct {
	SectionID string   `json:"section_id" validate:"required"`
	Title     string   `json:"title" validate:"required,max=240"`
	Abstract  string   `json:"abstract" validate:"required,max=2000"`
	Markdown  string   `json:"markdown" validate:"required"`
	Tags      []string `json:"tags" validate:"max=12"`
}

func (h *Handlers) CreateManuscript(c *gin.Context) {
	actor, _ := middleware.Actor(c)
	request, valid := decodeRequest[createManuscriptRequest](h, c, "manuscript fields are invalid")
	if !valid {
		return
	}
	created, err := h.manuscripts.Create(c.Request.Context(), application.CreateManuscriptInput{AuthorID: actor.ID, SectionID: request.SectionID, Title: request.Title, Abstract: request.Abstract, Markdown: request.Markdown, Tags: request.Tags})
	respond(c, http.StatusCreated, created, err)
}

type submitRequest struct {
	OriginalWork    bool   `json:"original_work"`
	AuthorApproved  bool   `json:"author_approved"`
	EthicsCleared   bool   `json:"ethics_cleared"`
	DeclarationText string `json:"declaration_text" validate:"required"`
	ExpectedVersion int64  `json:"expected_version" validate:"required,min=1"`
}

func (h *Handlers) SubmitManuscript(c *gin.Context) {
	actor, _ := middleware.Actor(c)
	request, valid := decodeRequest[submitRequest](h, c, "submission declaration is invalid")
	if !valid {
		return
	}
	snapshot, err := h.manuscripts.Submit(c.Request.Context(), application.SubmitManuscriptInput{ManuscriptID: c.Param("id"), AuthorID: actor.ID, ExpectedVersion: request.ExpectedVersion, IdempotencyKey: c.GetHeader("Idempotency-Key"), Declaration: manuscript.Declaration{OriginalWork: request.OriginalWork, AuthorApproved: request.AuthorApproved, EthicsCleared: request.EthicsCleared, Text: request.DeclarationText}})
	respond(c, http.StatusOK, snapshot, err)
}

func (h *Handlers) ListManuscripts(c *gin.Context) {
	actor, _ := middleware.Actor(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "25"))
	page, err := h.manuscripts.List(c.Request.Context(), actor, application.ManuscriptFilter{AuthorID: c.Query("author_id"), Section: c.Query("section"), Status: manuscript.Status(c.Query("status")), Sort: c.DefaultQuery("sort", "updated_at"), Cursor: c.Query("cursor"), Limit: limit})
	respond(c, http.StatusOK, page, err)
}

type assignmentRequest struct {
	ReviewerID string    `json:"reviewer_id" validate:"required"`
	DueAt      time.Time `json:"due_at" validate:"required"`
}

func (h *Handlers) Assign(c *gin.Context) {
	actor, _ := middleware.Actor(c)
	request, valid := decodeRequest[assignmentRequest](h, c, "assignment fields are invalid")
	if !valid {
		return
	}
	assignment, err := h.reviews.Assign(c.Request.Context(), c.Param("id"), request.ReviewerID, actor.ID, request.DueAt)
	respond(c, http.StatusCreated, assignment, err)
}

type reportRequest struct {
	Recommendation review.Recommendation      `json:"recommendation" validate:"required"`
	Scores         []review.ScoreItem         `json:"scores" validate:"required,min=1"`
	Comments       []review.StructuredComment `json:"comments" validate:"required,min=1"`
	AnonymousNote  string                     `json:"anonymous_note"`
}

func (h *Handlers) SubmitReport(c *gin.Context) {
	actor, _ := middleware.Actor(c)
	request, valid := decodeRequest[reportRequest](h, c, "review report is invalid")
	if !valid {
		return
	}
	report, err := h.reviews.SubmitReport(c.Request.Context(), c.Param("id"), actor.ID, request.Recommendation, request.Scores, request.Comments, request.AnonymousNote)
	respond(c, http.StatusCreated, presentSubmittedReport(report), err)
}

func presentSubmittedReport(report review.Report) review.Report {
	return report.PublicView()
}

func (h *Handlers) SearchPublications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	result, err := h.publications.Search(c.Request.Context(), application.PublicationFilter{Query: c.Query("query"), SectionID: c.Query("section"), Tag: c.Query("tag"), Sort: c.DefaultQuery("sort", "published_at"), Page: page, PageSize: pageSize})
	respond(c, http.StatusOK, result, err)
}

func decodeRequest[T any](handlers *Handlers, c *gin.Context, invalidMessage string) (T, bool) {
	var request T
	decodeErr := c.ShouldBindJSON(&request)
	if decodeErr == nil {
		decodeErr = handlers.validate.Struct(request)
	}
	if decodeErr == nil {
		return request, true
	}
	writeError(c, sharedValidation(invalidMessage))
	return request, false
}

func respond(c *gin.Context, status int, payload any, err error) {
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(status, payload)
}

func sharedValidation(message string) error {
	return &shared.DomainError{Code: "VALIDATION_FAILED", Message: message, Cause: shared.ErrValidation}
}
