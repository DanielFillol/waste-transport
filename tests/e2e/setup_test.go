package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/danielfillol/waste/internal/api"
	"github.com/danielfillol/waste/internal/api/handler"
	"github.com/danielfillol/waste/internal/config"
	"github.com/danielfillol/waste/internal/domain/entity"
	"github.com/danielfillol/waste/internal/infra/database"
	"github.com/danielfillol/waste/internal/infra/repository"
	authUC "github.com/danielfillol/waste/internal/usecase/auth"
	financialUC "github.com/danielfillol/waste/internal/usecase/financial"
	opsUC "github.com/danielfillol/waste/internal/usecase/operations"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

var (
	srv *httptest.Server
	db  *gorm.DB
	cfg *config.Config
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	_ = godotenv.Load("../../.env.test")

	cfg = config.Load()
	db = database.Connect(cfg)

	seedDomain(db)

	engine := buildEngine()
	srv = httptest.NewServer(engine)
	defer srv.Close()

	code := m.Run()

	cleanDB(db)
	os.Exit(code)
}

func buildEngine() *gin.Engine {
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

	authUseCase := authUC.NewUseCase(tenantRepo, userRepo, cfg)
	financialUseCase := financialUC.NewUseCase(financialRepo, collectRepo)
	alertUseCase := opsUC.NewAlertUseCase(alertRepo)
	dashboardUseCase := opsUC.NewDashboardUseCase(collectRepo, alertRepo, financialRepo, financialUseCase)

	authH := handler.NewAuthHandler(authUseCase)
	genH := handler.NewGeneratorHandler(generatorRepo)
	recH := handler.NewReceiverHandler(receiverRepo, alertUseCase)
	drvH := handler.NewDriverHandler(driverRepo, alertUseCase)
	trkH := handler.NewTruckHandler(truckRepo)
	rteH := handler.NewRouteHandler(routeRepo, driverRepo, collectRepo)
	colH := handler.NewCollectHandler(collectRepo)
	domH := handler.NewDomainHandler(domainRepo)
	finH := handler.NewFinancialHandler(financialUseCase)
	alrH := handler.NewAlertHandler(alertUseCase)
	impH := handler.NewImportHandler(generatorRepo, receiverRepo, driverRepo, collectRepo)
	dashH := handler.NewDashboardHandler(dashboardUseCase)
	audH := handler.NewAuditHandler(auditRepo)

	engine := gin.New()
	engine.Use(gin.Recovery())

	r := api.NewRouter(authH, genH, recH, drvH, trkH, rteH, colH, domH, finH, alrH, impH, dashH, audH, auditRepo, cfg)
	r.Setup(engine)
	return engine
}

func seedDomain(db *gorm.DB) {
	materials := []entity.Material{
		{ID: 1, Name: "Papel/Papelão", Description: "Resíduos de papel e papelão"},
		{ID: 2, Name: "Plástico", Description: "Resíduos plásticos"},
		{ID: 3, Name: "Metal", Description: "Resíduos metálicos"},
		{ID: 4, Name: "Vidro", Description: "Resíduos de vidro"},
		{ID: 5, Name: "Orgânico", Description: "Resíduos orgânicos"},
		{ID: 6, Name: "Eletrônico", Description: "Resíduos eletroeletrônicos"},
	}
	for _, m := range materials {
		db.FirstOrCreate(&m, "id = ?", m.ID)
	}

	packagings := []entity.Packaging{
		{ID: 1, Name: "Bag", Type: "flexible", Volume: 1000},
		{ID: 2, Name: "Tambor 200L", Type: "rigid", Volume: 200},
		{ID: 3, Name: "Contêiner", Type: "rigid", Volume: 5000},
		{ID: 4, Name: "Granel", Type: "bulk", Volume: 0},
	}
	for _, p := range packagings {
		db.FirstOrCreate(&p, "id = ?", p.ID)
	}

	treatments := []entity.Treatment{
		{ID: 1, Name: "Reciclagem", Description: "Reciclagem de materiais"},
		{ID: 2, Name: "Coprocessamento", Description: "Coprocessamento em fornos de cimento"},
		{ID: 3, Name: "Aterro sanitário", Description: "Disposição em aterro"},
		{ID: 4, Name: "Incineração", Description: "Destruição térmica"},
	}
	for _, t := range treatments {
		db.FirstOrCreate(&t, "id = ?", t.ID)
	}

	ufs := []entity.UF{
		{ID: 1, Name: "São Paulo", Code: "SP"},
		{ID: 2, Name: "Rio de Janeiro", Code: "RJ"},
	}
	for _, u := range ufs {
		db.FirstOrCreate(&u, "id = ?", u.ID)
	}

	cities := []entity.City{
		{ID: 1, Name: "São Paulo", UFID: 1},
		{ID: 2, Name: "Campinas", UFID: 1},
		{ID: 3, Name: "Rio de Janeiro", UFID: 2},
	}
	for _, c := range cities {
		db.FirstOrCreate(&c, "id = ?", c.ID)
	}
}

func cleanDB(db *gorm.DB) {
	tables := []string{
		"audit_logs",
		"invoice_items", "invoices", "pricing_rules",
		"personnel_costs", "truck_costs",
		"collects", "driver_routes", "routes",
		"trucks", "drivers", "receivers", "generators",
		"alerts", "users", "tenants",
	}
	for _, t := range tables {
		db.Exec(fmt.Sprintf("DELETE FROM %s", t))
	}
}

// url builds a full URL for the test server
func url(path string) string {
	return fmt.Sprintf("%s%s", srv.URL, path)
}

// doRequest executes an HTTP request with optional JSON body and auth header
func doRequest(t *testing.T, method, path string, body interface{}, token string) *http.Response {
	t.Helper()
	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request body: %v", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, url(path), buf)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

// decodeJSON decodes response body into v
func decodeJSON(t *testing.T, resp *http.Response, v interface{}) {
	t.Helper()
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v", err)
	}
}

// mustStatus asserts that the response has the expected status code
func mustStatus(t *testing.T, resp *http.Response, expected int) {
	t.Helper()
	if resp.StatusCode != expected {
		var body map[string]interface{}
		decodeJSON(t, resp, &body)
		t.Fatalf("expected status %d, got %d — body: %v", expected, resp.StatusCode, body)
	}
}

// setupTenant registers a tenant and returns the token + tenant info
func setupTenant(t *testing.T, name string) (token string, slug string) {
	t.Helper()
	resp := doRequest(t, "POST", "/v1/auth/tenants", map[string]string{"name": name}, "")
	mustStatus(t, resp, http.StatusCreated)

	var result map[string]interface{}
	decodeJSON(t, resp, &result)
	token = result["token"].(string)
	slug = result["tenant"].(map[string]interface{})["slug"].(string)
	return token, slug
}
