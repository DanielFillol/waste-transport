package main

import (
	"fmt"
	"log"

	"github.com/danielfillol/waste/internal/api"
	"github.com/danielfillol/waste/internal/api/handler"
	"github.com/danielfillol/waste/internal/config"
	"github.com/danielfillol/waste/internal/infra/database"
	"github.com/danielfillol/waste/internal/infra/repository"
	authUC "github.com/danielfillol/waste/internal/usecase/auth"
	financialUC "github.com/danielfillol/waste/internal/usecase/financial"
	opsUC "github.com/danielfillol/waste/internal/usecase/operations"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	cfg := config.Load()
	db := database.Connect(cfg)

	// Repositories
	tenantRepo := repository.NewTenantRepository(db)
	userRepo := repository.NewUserRepository(db)
	generatorRepo := repository.NewGeneratorRepository(db)
	receiverRepo := repository.NewReceiverRepository(db)
	driverRepo := repository.NewDriverRepository(db)
	truckRepo := repository.NewTruckRepository(db)
	routeRepo := repository.NewRouteRepository(db)
	collectRepo := repository.NewCollectRepository(db)
	domainRepo := repository.NewDomainRepository(db)
	financialRepo := repository.NewFinancialRepository(db)
	alertRepo := repository.NewAlertRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// Use cases
	authUseCase := authUC.NewUseCase(tenantRepo, userRepo, cfg)
	financialUseCase := financialUC.NewUseCase(financialRepo, collectRepo)
	alertUseCase := opsUC.NewAlertUseCase(alertRepo)
	dashboardUseCase := opsUC.NewDashboardUseCase(collectRepo, alertRepo, financialRepo, financialUseCase)

	// Handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	generatorHandler := handler.NewGeneratorHandler(generatorRepo)
	receiverHandler := handler.NewReceiverHandler(receiverRepo, alertUseCase)
	driverHandler := handler.NewDriverHandler(driverRepo, alertUseCase)
	alertHandler := handler.NewAlertHandler(alertUseCase)
	dashboardHandler := handler.NewDashboardHandler(dashboardUseCase)
	importHandler := handler.NewImportHandler(generatorRepo, receiverRepo, driverRepo, collectRepo, truckRepo, routeRepo)
	truckHandler := handler.NewTruckHandler(truckRepo)
	routeHandler := handler.NewRouteHandler(routeRepo, driverRepo, collectRepo)
	collectHandler := handler.NewCollectHandler(collectRepo)
	domainHandler := handler.NewDomainHandler(domainRepo)
	financialHandler := handler.NewFinancialHandler(financialUseCase)
	auditHandler := handler.NewAuditHandler(auditRepo)

	// Engine
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.Default()

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Routes
	router := api.NewRouter(
		authHandler,
		generatorHandler,
		receiverHandler,
		driverHandler,
		truckHandler,
		routeHandler,
		collectHandler,
		domainHandler,
		financialHandler,
		alertHandler,
		importHandler,
		dashboardHandler,
		auditHandler,
		auditRepo,
		cfg,
	)
	router.Setup(engine)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("waste API listening on %s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
