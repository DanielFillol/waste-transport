package handler

import (
	"net/http"
	"strconv"

	"github.com/danielfillol/waste/internal/infra/repository"
	"github.com/gin-gonic/gin"
)

type DomainHandler struct {
	repo *repository.DomainRepository
}

func NewDomainHandler(repo *repository.DomainRepository) *DomainHandler {
	return &DomainHandler{repo: repo}
}

func (h *DomainHandler) ListMaterials(c *gin.Context) {
	items, err := h.repo.ListMaterials()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *DomainHandler) ListPackagings(c *gin.Context) {
	items, err := h.repo.ListPackagings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *DomainHandler) ListTreatments(c *gin.Context) {
	items, err := h.repo.ListTreatments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *DomainHandler) ListUFs(c *gin.Context) {
	items, err := h.repo.ListUFs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *DomainHandler) ListCities(c *gin.Context) {
	var ufID *uint
	if v := c.Query("uf_id"); v != "" {
		n, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			u := uint(n)
			ufID = &u
		}
	}
	items, err := h.repo.ListCities(ufID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}
