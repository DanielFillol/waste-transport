package api

import (
	"github.com/danielfillol/waste/internal/api/handler"
	"github.com/danielfillol/waste/internal/api/middleware"
	"github.com/danielfillol/waste/internal/config"
	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/gin-gonic/gin"
)

type Router struct {
	auth      *handler.AuthHandler
	generator *handler.GeneratorHandler
	receiver  *handler.ReceiverHandler
	driver    *handler.DriverHandler
	truck     *handler.TruckHandler
	route     *handler.RouteHandler
	collect   *handler.CollectHandler
	domain    *handler.DomainHandler
	financial *handler.FinancialHandler
	alert     *handler.AlertHandler
	importer  *handler.ImportHandler
	dashboard *handler.DashboardHandler
	audit     *handler.AuditHandler
	auditRepo *repository.AuditRepository
	cfg       *config.Config
}

func NewRouter(
	auth *handler.AuthHandler,
	generator *handler.GeneratorHandler,
	receiver *handler.ReceiverHandler,
	driver *handler.DriverHandler,
	truck *handler.TruckHandler,
	route *handler.RouteHandler,
	collect *handler.CollectHandler,
	domain *handler.DomainHandler,
	financial *handler.FinancialHandler,
	alert *handler.AlertHandler,
	importer *handler.ImportHandler,
	dashboard *handler.DashboardHandler,
	audit *handler.AuditHandler,
	auditRepo *repository.AuditRepository,
	cfg *config.Config,
) *Router {
	return &Router{
		auth:      auth,
		generator: generator,
		receiver:  receiver,
		driver:    driver,
		truck:     truck,
		route:     route,
		collect:   collect,
		domain:    domain,
		financial: financial,
		alert:     alert,
		importer:  importer,
		dashboard: dashboard,
		audit:     audit,
		auditRepo: auditRepo,
		cfg:       cfg,
	}
}

func (r *Router) Setup(engine *gin.Engine) {
	engine.Use(gin.Recovery())

	v1 := engine.Group("/v1")

	// Public
	v1.POST("/auth/tenants", r.auth.RegisterTenant)
	v1.POST("/auth/login", r.auth.Login)

	// Protected
	protected := v1.Group("")
	protected.Use(middleware.Auth(r.cfg))
	protected.Use(middleware.Audit(r.auditRepo))

	protected.GET("/me", r.auth.Me)
	protected.POST("/auth/refresh", r.auth.Refresh)
	protected.GET("/dashboard", r.dashboard.Get)

	// Users (admin only)
	users := protected.Group("/users")
	users.Use(middleware.AdminOnly())
	{
		users.GET("", r.auth.ListUsers)
		users.POST("", r.auth.CreateUser)
		users.PUT("/:id", r.auth.UpdateUser)
		users.DELETE("/:id", r.auth.DeleteUser)
	}

	// Generators
	gen := protected.Group("/generators")
	{
		gen.GET("", r.generator.List)
		gen.POST("", r.generator.Create)
		gen.POST("/import", r.importer.ImportGenerators)
		gen.POST("/import-delete", r.importer.ImportDeleteGenerators)
		gen.GET("/:id", r.generator.Get)
		gen.PATCH("/:id", r.generator.Update)
		gen.DELETE("/:id", r.generator.Delete)
	}

	// Receivers
	rec := protected.Group("/receivers")
	{
		rec.GET("", r.receiver.List)
		rec.POST("", r.receiver.Create)
		rec.POST("/import", r.importer.ImportReceivers)
		rec.POST("/import-delete", r.importer.ImportDeleteReceivers)
		rec.GET("/:id", r.receiver.Get)
		rec.PATCH("/:id", r.receiver.Update)
		rec.DELETE("/:id", r.receiver.Delete)
	}

	// Domain (read-only)
	dom := protected.Group("/domain")
	{
		dom.GET("/materials", r.domain.ListMaterials)
		dom.GET("/packagings", r.domain.ListPackagings)
		dom.GET("/treatments", r.domain.ListTreatments)
		dom.GET("/ufs", r.domain.ListUFs)
		dom.GET("/cities", r.domain.ListCities)
	}

	// Drivers
	drv := protected.Group("/drivers")
	{
		drv.GET("", r.driver.List)
		drv.POST("", r.driver.Create)
		drv.POST("/import", r.importer.ImportDrivers)
		drv.POST("/import-delete", r.importer.ImportDeleteDrivers)
		drv.GET("/:id", r.driver.Get)
		drv.PATCH("/:id", r.driver.Update)
		drv.DELETE("/:id", r.driver.Delete)
	}

	// Trucks
	trk := protected.Group("/trucks")
	{
		trk.GET("", r.truck.List)
		trk.POST("", r.truck.Create)
		trk.POST("/import", r.importer.ImportTrucks)
		trk.POST("/import-delete", r.importer.ImportDeleteTrucks)
		trk.GET("/:id", r.truck.Get)
		trk.PATCH("/:id", r.truck.Update)
		trk.DELETE("/:id", r.truck.Delete)
	}

	// Routes
	rts := protected.Group("/routes")
	{
		rts.GET("", r.route.List)
		rts.POST("", r.route.Create)
		rts.POST("/import", r.importer.ImportRoutes)
		rts.POST("/import-delete", r.importer.ImportDeleteRoutes)
		rts.GET("/:id", r.route.Get)
		rts.PATCH("/:id", r.route.Update)
		rts.DELETE("/:id", r.route.Delete)
		rts.POST("/:id/generate-collects", r.route.GenerateCollects)
	}

	// Collects
	col := protected.Group("/collects")
	{
		col.GET("", r.collect.List)
		col.POST("", r.collect.Create)
		col.GET("/:id", r.collect.Get)
		col.PATCH("/:id", r.collect.Update)
		col.DELETE("/:id", r.collect.Delete)
		col.POST("/import", r.importer.ImportCollects)
		col.POST("/import-delete", r.importer.ImportDeleteCollects)
		col.POST("/bulk-status", r.collect.BulkStatus)
		col.POST("/bulk-cancel", r.collect.BulkCancel)
		col.POST("/bulk-assign-route", r.collect.BulkAssignRoute)
	}

	// Alerts
	alr := protected.Group("/alerts")
	{
		alr.GET("", r.alert.List)
		alr.PATCH("/read-all", r.alert.MarkAllRead)
		alr.PATCH("/:id/read", r.alert.MarkRead)
	}

	// Financial
	fin := protected.Group("/financial")
	{
		// Pricing Rules
		fin.GET("/pricing-rules", r.financial.ListPricingRules)
		fin.POST("/pricing-rules", r.financial.CreatePricingRule)
		fin.GET("/pricing-rules/:id", r.financial.GetPricingRule)
		fin.PUT("/pricing-rules/:id", r.financial.UpdatePricingRule)
		fin.DELETE("/pricing-rules/:id", r.financial.DeletePricingRule)

		// Financial Summary
		fin.GET("/summary", r.financial.Summary)

		// Invoices
		fin.GET("/invoices", r.financial.ListInvoices)
		fin.POST("/invoices/generate", r.financial.GenerateInvoice)
		fin.GET("/invoices/:id", r.financial.GetInvoice)
		fin.PATCH("/invoices/:id/issue", r.financial.IssueInvoice)
		fin.PATCH("/invoices/:id/paid", r.financial.MarkInvoicePaid)

		// Truck Costs
		fin.GET("/truck-costs", r.financial.ListTruckCosts)
		fin.POST("/truck-costs", r.financial.CreateTruckCost)
		fin.GET("/truck-costs/:id", r.financial.GetTruckCost)
		fin.PUT("/truck-costs/:id", r.financial.UpdateTruckCost)
		fin.DELETE("/truck-costs/:id", r.financial.DeleteTruckCost)

		// Personnel Costs
		fin.GET("/personnel-costs", r.financial.ListPersonnelCosts)
		fin.POST("/personnel-costs", r.financial.CreatePersonnelCost)
		fin.GET("/personnel-costs/:id", r.financial.GetPersonnelCost)
		fin.PUT("/personnel-costs/:id", r.financial.UpdatePersonnelCost)
		fin.DELETE("/personnel-costs/:id", r.financial.DeletePersonnelCost)
	}

	// Audit Logs
	protected.GET("/audit-logs", r.audit.List)
}
