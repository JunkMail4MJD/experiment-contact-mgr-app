package handlers

import (
	"contactmanager/models"
	"contactmanager/service"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type RESTHandler struct {
	service *service.ContactService
}

func NewRESTHandler(service *service.ContactService) *RESTHandler {
	return &RESTHandler{service: service}
}

func (h *RESTHandler) SetupRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")
	{
		contacts := api.Group("/contacts")
		{
			contacts.GET("", h.ListContacts)
			contacts.POST("", h.CreateContact)
			contacts.GET("/:id", h.GetContact)
			contacts.PUT("/:id", h.UpdateContact)
			contacts.DELETE("/:id", h.DeleteContact)
			contacts.POST("/bulk", h.BulkCreateContacts)
		}
	}
}

type ListContactsResponse struct {
	Contacts   []*models.Contact `json:"contacts"`
	Pagination PaginationInfo    `json:"pagination"`
}

type PaginationInfo struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	Total       int  `json:"total"`
	TotalPages  int  `json:"totalPages"`
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
}

func (h *RESTHandler) ListContacts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	tag := c.Query("tag")

	opts := models.ListOptions{
		Page:   page,
		Limit:  limit,
		Search: search,
		Tag:    tag,
	}

	contacts, total, err := h.service.ListContacts(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + limit - 1) / limit

	response := ListContactsResponse{
		Contacts: contacts,
		Pagination: PaginationInfo{
			Page:        page,
			Limit:       limit,
			Total:       total,
			TotalPages:  totalPages,
			HasNext:     page < totalPages,
			HasPrevious: page > 1,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *RESTHandler) CreateContact(c *gin.Context) {
	var input models.ContactInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := h.service.CreateContact(&input)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, gin.H{"error": "Contact with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, contact)
}

func (h *RESTHandler) GetContact(c *gin.Context) {
	id := c.Param("id")

	contact, err := h.service.GetContact(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, contact)
}

func (h *RESTHandler) UpdateContact(c *gin.Context) {
	id := c.Param("id")

	var input models.ContactInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contact, err := h.service.UpdateContact(id, &input)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists for another contact"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, contact)
}

func (h *RESTHandler) DeleteContact(c *gin.Context) {
	id := c.Param("id")

	err := h.service.DeleteContact(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Contact not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *RESTHandler) BulkCreateContacts(c *gin.Context) {
	var request struct {
		Contacts []*models.ContactInput `json:"contacts" binding:"required,min=1,max=100"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	contacts, errors := h.service.BulkCreateContacts(request.Contacts)

	var created []*models.Contact
	var bulkErrors []map[string]interface{}

	for i, contact := range contacts {
		if errors[i] != nil {
			bulkErrors = append(bulkErrors, map[string]interface{}{
				"index": i,
				"error": errors[i].Error(),
			})
		} else if contact != nil {
			created = append(created, contact)
		}
	}

	response := map[string]interface{}{
		"created": created,
		"errors":  bulkErrors,
	}

	c.JSON(http.StatusCreated, response)
}
