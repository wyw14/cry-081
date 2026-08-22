package bootstrap

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-081/internal/application"
	"github.com/wyw14/cry-081/internal/config"
	"github.com/wyw14/cry-081/internal/domain/manuscript"
	"github.com/wyw14/cry-081/internal/platform/clock"
	"github.com/wyw14/cry-081/internal/platform/idempotency"
	"github.com/wyw14/cry-081/internal/platform/outbox"
	"github.com/wyw14/cry-081/internal/repository/memory"
	"github.com/wyw14/cry-081/internal/repository/postgres"
	httptransport "github.com/wyw14/cry-081/internal/transport/http"
	"go.uber.org/zap"
)

type App struct {
	Config   config.Config
	Logger   *zap.Logger
	Router   *gin.Engine
	database *postgres.Database
	ready    atomic.Bool
}

type stores struct {
	users        application.UserRepository
	tokens       application.TokenRepository
	manuscripts  application.ManuscriptRepository
	assignments  application.AssignmentRepository
	reports      application.ReportRepository
	decisions    application.DecisionRepository
	issues       application.IssueRepository
	articles     application.ArticleRepository
	audits       application.AuditRepository
	outbox       outbox.Store
	replays      idempotency.Store
	transactions application.TransactionManager
}

func Build(ctx context.Context, cfg config.Config, logger *zap.Logger) (*App, error) {
	app := &App{Config: cfg, Logger: logger}
	storeSet, err := app.buildStores(ctx, cfg)
	if err != nil {
		return nil, err
	}
	ids := application.NewSequentialIDs("cry081")
	systemClock := clock.System()
	identityService := application.NewIdentityService(storeSet.users, storeSet.tokens, application.BcryptHasher{}, application.LocalTokenIssuer{AccessTTL: cfg.Tokens.AccessTTL}, ids, systemClock, cfg.Tokens.RefreshTTL)
	manuscriptService := application.NewManuscriptService(storeSet.manuscripts, storeSet.users, storeSet.audits, storeSet.transactions, storeSet.outbox, storeSet.replays, ids, systemClock, manuscript.DefaultFormatPolicy())
	reviewService := application.NewReviewService(storeSet.manuscripts, storeSet.assignments, storeSet.reports, storeSet.decisions, storeSet.users, storeSet.audits, storeSet.transactions, storeSet.outbox, ids, systemClock)
	publicationService := application.NewPublicationService(storeSet.manuscripts, storeSet.issues, storeSet.articles, storeSet.users, storeSet.audits, storeSet.transactions, ids, systemClock)
	reportingService := application.NewReportingService(storeSet.manuscripts, storeSet.assignments, storeSet.audits)
	handlers := httptransport.NewHandlers(identityService, manuscriptService, reviewService, publicationService, reportingService)
	app.Router = httptransport.NewRouter(handlers, storeSet.users, app, cfg.HTTP.AllowedOrigin, cfg.HTTP.RequestsPerMin, cfg.HTTP.RequestTimeout)
	app.ready.Store(true)
	return app, nil
}

func (a *App) buildStores(ctx context.Context, cfg config.Config) (stores, error) {
	if cfg.Storage.Driver == "postgres" {
		database, err := postgres.Open(ctx, cfg.Storage.DatabaseURL)
		if err != nil {
			return stores{}, err
		}
		a.database = database
		return stores{
			users: postgres.NewUserStore(database), tokens: postgres.NewTokenStore(database),
			manuscripts: postgres.NewManuscriptStore(database), assignments: postgres.NewAssignmentStore(database),
			reports: postgres.NewReportStore(database), decisions: postgres.NewDecisionStore(database),
			issues: postgres.NewIssueStore(database), articles: postgres.NewArticleStore(database),
			audits: postgres.NewAuditStore(database), outbox: postgres.NewOutboxStore(database), replays: postgres.NewIdempotencyStore(database), transactions: database,
		}, nil
	}
	identityStore := memory.NewIdentityStore()
	reviewStore := memory.NewReviewStore()
	publicationStore := memory.NewPublicationStore()
	return stores{
		users: identityStore, tokens: memory.NewTokenStore(identityStore), manuscripts: memory.NewManuscriptStore(),
		assignments: reviewStore, reports: memory.NewReportStore(reviewStore), decisions: memory.NewDecisionStore(reviewStore),
		issues: memory.NewIssueStore(publicationStore), articles: memory.NewArticleStore(publicationStore),
		audits: memory.NewAuditStore(), outbox: outbox.NewMemoryStore(), replays: idempotency.NewMemory(), transactions: application.DirectTransactionManager{},
	}, nil
}

func (a *App) Ready() bool { return a.ready.Load() }

func (a *App) Close(ctx context.Context) error {
	a.ready.Store(false)
	if a.database != nil {
		done := make(chan struct{})
		go func() {
			a.database.Close()
			close(done)
		}()
		select {
		case <-ctx.Done():
			return fmt.Errorf("database close: %w", ctx.Err())
		case <-done:
		}
	}
	return nil
}

func DevelopmentLogger(environment string) (*zap.Logger, error) {
	if environment == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
