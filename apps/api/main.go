package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/ai"
	"github.com/zzm/opcv2/internal/analysis"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/content"
	"github.com/zzm/opcv2/internal/copilot"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/dashboard"
	"github.com/zzm/opcv2/internal/enterprise"
	"github.com/zzm/opcv2/internal/geo"
	"github.com/zzm/opcv2/internal/growth"
	"github.com/zzm/opcv2/internal/home"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/learning"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/notifications"
	"github.com/zzm/opcv2/internal/platform/aiprovider"
	"github.com/zzm/opcv2/internal/platform/config"
	"github.com/zzm/opcv2/internal/platform/health"
	"github.com/zzm/opcv2/internal/platform/httpserver"
	"github.com/zzm/opcv2/internal/platform/migrations"
	"github.com/zzm/opcv2/internal/platform/postgres"
	"github.com/zzm/opcv2/internal/platform/projectprovider"
	"github.com/zzm/opcv2/internal/platform/rediscache"
	"github.com/zzm/opcv2/internal/platform/taskqueue"
	"github.com/zzm/opcv2/internal/projects"
	taskfiles "github.com/zzm/opcv2/internal/projects/files"
	"github.com/zzm/opcv2/internal/sandbox"
	"github.com/zzm/opcv2/internal/support"
	"github.com/zzm/opcv2/internal/tasks"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	db, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("open postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if cfg.AutoMigrate {
		result, err := migrations.Run(ctx, db, cfg.MigrationsPath)
		if err != nil {
			logger.Error("run migrations", "path", cfg.MigrationsPath, "error", err)
			os.Exit(1)
		}
		logger.Info("migrations complete", "path", cfg.MigrationsPath, "current_version", result.CurrentVersion, "applied_count", len(result.Applied))
	}

	redisClient := rediscache.NewClient(cfg.RedisAddr)
	defer func() { _ = redisClient.Close() }()

	userRepository := auth.NewPostgresUserRepository(db)
	codeStore := auth.NewRedisCodeStore(redisClient.Client)
	sessionStore := auth.NewRedisSessionStore(redisClient.Client)
	tokenManager := auth.NewJWTManager(cfg.JWTSecret, 15*time.Minute)
	var smsProvider auth.SMSProvider
	switch cfg.SMSProvider {
	case "development":
		smsProvider = auth.NewDevelopmentSMSProvider(cfg.SMSDevCode)
	case "disabled":
		smsProvider = auth.DisabledSMSProvider{}
	default:
		logger.Error("unsupported SMS provider", "provider", cfg.SMSProvider)
		os.Exit(1)
	}
	authService := auth.NewService(auth.Dependencies{
		Codes:        codeStore,
		SMS:          smsProvider,
		Users:        userRepository,
		Tokens:       tokenManager,
		Sessions:     sessionStore,
		GenerateCode: func() (string, error) { return cfg.SMSDevCode, nil },
	})
	authHTTP := auth.NewHTTPHandler(authService, tokenManager, cfg.Environment == "production")
	accountRepository := account.NewPostgresRepository(db)
	accountService := account.NewService(accountRepository)
	accountHTTP := account.NewHTTPHandler(accountService)
	notificationsRepository := notifications.NewPostgresRepository(db)
	notificationsService := notifications.NewService(notificationsRepository)
	notificationsHTTP := notifications.NewHTTPHandler(notificationsService)
	membershipRepository := membership.NewPostgresRepository(db)
	membershipService := membership.NewService(membershipRepository)
	membershipHTTP := membership.NewHTTPHandler(membershipService)
	aiRepository := ai.NewPostgresRepository(db)
	aiProvider, err := newAIProvider(cfg)
	if err != nil {
		logger.Error("configure ai provider", "provider", cfg.AIProvider, "error", err)
		os.Exit(1)
	}
	aiService := ai.NewService(aiRepository, aiProvider, ai.Config{
		Provider: cfg.AIProvider,
		Model:    cfg.AIModel,
		Logger:   logger,
	})
	analysisRepository := analysis.NewPostgresRepository(db)
	analysisService := analysis.NewService(analysisRepository, aiService)
	analysisHTTP := analysis.NewHTTPHandler(analysisService)
	projectsRepository := projects.NewPostgresRepository(db)
	projectProviders, err := projectprovider.New(cfg, projectprovider.WithRetrievalProvider(projectsRepository))
	if err != nil {
		logger.Error("configure project providers", "error", err)
		os.Exit(1)
	}
	projectProviders.Files.SetMetadataStore(projectsRepository)
	projectsService := projects.NewService(
		projectsRepository,
		aiService,
		projects.WithProfileContextProvider(accountService),
		projects.WithFeaturePaywallEnabled(cfg.FeaturePaywallEnabled),
		projects.WithProjectMatchQueue(taskqueue.NewClient(cfg.RedisAddr)),
		projects.WithProjectFileManager(projectProviders.Files),
		projects.WithProjectRetrievalProvider(projectProviders.Retrieval),
		projects.WithProjectResearchService(projectProviders.Research),
		projects.WithProjectUsageConsumer(membershipService),
	)
	projectsHTTP := projects.NewHTTPHandler(projectsService)
	leadsRepository := leads.NewPostgresRepository(db)
	leadProvider, err := newLeadProvider(cfg)
	if err != nil {
		logger.Error("configure lead provider", "provider", cfg.LeadProvider, "error", err)
		os.Exit(1)
	}
	leadsService := leads.NewService(
		leadsRepository,
		membershipService,
		taskqueue.NewClient(cfg.RedisAddr),
		leadProvider,
	)
	leadsHTTP := leads.NewHTTPHandler(leadsService)
	crmRepository := crm.NewPostgresRepository(db)
	crmService := crm.NewService(crmRepository, aiService)
	crmHTTP := crm.NewHTTPHandler(crmService)
	contentRepository := content.NewPostgresRepository(db)
	contentService := content.NewService(contentRepository, content.WithGenerator(aiService))
	contentHTTP := content.NewHTTPHandler(contentService)
	supportRepository := support.NewPostgresRepository(db)
	supportService := support.NewService(supportRepository)
	supportHTTP := support.NewHTTPHandler(supportService)
	sandboxRepository := sandbox.NewPostgresRepository(db)
	tasksRepository := tasks.NewPostgresRepository(db)
	taskAttachmentStorage, err := taskfiles.NewLocalStorage(filepath.Join(cfg.ProjectFileStoragePath, "task-attachments"))
	if err != nil {
		logger.Error("configure task attachment storage", "error", err)
		os.Exit(1)
	}
	tasksService := tasks.NewService(
		tasksRepository,
		tasks.WithMembershipProvider(membershipService),
		tasks.WithTaskGenerator(aiService),
		tasks.WithTaskAttachmentStorage(taskAttachmentStorage, taskfiles.DevelopmentScanner{}, cfg.JWTSecret),
	)
	sandboxService := sandbox.NewService(
		sandboxRepository,
		aiService,
		sandbox.WithQuotaConsumer(membershipService),
		sandbox.WithProfileContextProvider(accountService),
		sandbox.WithQueue(taskqueue.NewClient(cfg.RedisAddr)),
		sandbox.WithTaskCreator(tasksService),
	)
	sandboxHTTP := sandbox.NewHTTPHandler(sandboxService)
	tasksHTTP := tasks.NewHTTPHandler(tasksService)
	dashboardRepository := dashboard.NewPostgresRepository(db)
	dashboardService := dashboard.NewService(dashboardRepository)
	dashboardHTTP := dashboard.NewHTTPHandler(dashboardService)
	geoRepository := geo.NewPostgresRepository(db)
	geoService := geo.NewService(geoRepository, geo.WithQueue(taskqueue.NewClient(cfg.RedisAddr)))
	geoHTTP := geo.NewHTTPHandler(geoService)
	enterpriseRepository := enterprise.NewPostgresRepository(db)
	enterpriseService := enterprise.NewService(enterpriseRepository, crmService)
	enterpriseHTTP := enterprise.NewHTTPHandler(enterpriseService)
	growthRepository := growth.NewPostgresRepository(db)
	growthService := growth.NewService(growthRepository)
	growthHTTP := growth.NewHTTPHandler(growthService)
	competitorRepository := competitor.NewPostgresRepository(db)
	competitorService := competitor.NewService(
		competitorRepository,
		competitor.WithQuotaConsumer(membershipService),
		competitor.WithQueue(taskqueue.NewClient(cfg.RedisAddr)),
	)
	competitorHTTP := competitor.NewHTTPHandler(competitorService)
	homeService := home.NewService(home.Dependencies{
		Notifications: notificationsService,
		Membership:    membershipService,
		Tasks:         tasksService,
		Leads:         leadsService,
		Sandbox:       sandboxService,
		Competitor:    competitorService,
		CRM:           crmService,
	})
	homeHTTP := home.NewHTTPHandler(homeService)
	learningRepository := learning.NewPostgresRepository(db)
	learningService := learning.NewService(
		learningRepository,
		learning.WithGenerator(aiService),
		learning.WithProfileContextProvider(accountService),
	)
	learningHTTP := learning.NewHTTPHandler(learningService)
	copilotRepository := copilot.NewPostgresRepository(db)
	copilotService := copilot.NewServiceWithModels(
		copilotRepository,
		aiService,
		copilotModelOptions(cfg),
		copilot.WithQuotaConsumer(membershipService),
		copilot.WithToolExecutor(copilot.NewToolRegistry(tasksService, projectsService)),
		copilot.WithTaskContextProvider(tasksService),
		copilot.WithCompetitorContextProvider(competitorService),
		copilot.WithGrowthContextProvider(growthService),
		copilot.WithMonitoringContextProvider(competitorService),
		copilot.WithProjectContextProvider(projectsService),
		copilot.WithLearningContextProvider(learningService),
	)
	copilotHTTP := copilot.NewHTTPHandler(copilotService)

	checker := health.NewChecker(db, redisClient)
	server := &http.Server{
		Addr: cfg.HTTPAddr,
		Handler: httpserver.NewRouter(httpserver.HealthChecks{
			Ready:                  func() bool { return checker.Ready(context.Background()) },
			Logger:                 logger,
			AllowedOrigins:         cfg.CORSAllowedOrigins,
			ExpensiveEndpointLimit: cfg.ExpensiveEndpointLimit,
		}, httpserver.Handlers{
			Auth:          authHTTP,
			Account:       accountHTTP,
			Notifications: notificationsHTTP,
			Home:          homeHTTP,
			Membership:    membershipHTTP,
			Analysis:      analysisHTTP,
			Projects:      projectsHTTP,
			Leads:         leadsHTTP,
			CRM:           crmHTTP,
			Content:       contentHTTP,
			Support:       supportHTTP,
			Sandbox:       sandboxHTTP,
			Tasks:         tasksHTTP,
			Dashboard:     dashboardHTTP,
			Geo:           geoHTTP,
			Enterprise:    enterpriseHTTP,
			Growth:        growthHTTP,
			Competitor:    competitorHTTP,
			Learning:      learningHTTP,
			Copilot:       copilotHTTP,
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("api listening", "addr", cfg.HTTPAddr, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("serve api", "error", err)
			stop()
		}
	}()

	<-shutdownCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown api", "error", err)
	}
}

func newLeadProvider(cfg config.Config) (leads.LeadProvider, error) {
	switch cfg.LeadProvider {
	case "development":
		return leads.NewDevelopmentProvider(), nil
	case "tianyancha":
		return leads.NewTianyanchaProvider(leads.TianyanchaConfig{
			BaseURL: cfg.TianyanchaBaseURL,
			APIKey:  cfg.TianyanchaAPIKey,
			Timeout: time.Duration(cfg.TianyanchaTimeoutSeconds) * time.Second,
		}), nil
	default:
		return nil, errors.New("unsupported lead provider")
	}
}

func newAIProvider(cfg config.Config) (ai.Provider, error) {
	return aiprovider.New(cfg)
}

func copilotModelOptions(cfg config.Config) []copilot.ModelOption {
	if len(cfg.AIModelRoutes) == 0 {
		return []copilot.ModelOption{{
			Name:      displayModelName(cfg.AIModel),
			Value:     cfg.AIModel,
			Provider:  cfg.AIProvider,
			IsDefault: true,
		}}
	}
	options := make([]copilot.ModelOption, 0, len(cfg.AIModelRoutes))
	for index, route := range cfg.AIModelRoutes {
		options = append(options, copilot.ModelOption{
			Name:      displayModelName(route.Alias),
			Value:     route.Alias,
			Provider:  route.Provider,
			IsDefault: route.Alias == cfg.AIModel || (cfg.AIModel == "" && index == 0),
		})
	}
	return options
}

func displayModelName(alias string) string {
	switch alias {
	case "deepseek":
		return "DeepSeek"
	case "gpt-main":
		return "GPT-4o"
	case "development-model":
		return "Development"
	default:
		return alias
	}
}
