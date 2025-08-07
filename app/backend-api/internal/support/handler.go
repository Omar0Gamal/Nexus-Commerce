package support

import (
	"errors"
	"strconv"

	"backend-api/internal/shared/ginutil"
	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterCustomerRoutes(rg *gin.RouterGroup, requireAuth, requireCustomer gin.HandlerFunc) {
	tickets := rg.Group("/support/my-tickets")
	tickets.Use(requireAuth, requireCustomer)
	{
		tickets.POST("", h.CreateTicket)
		tickets.GET("", h.GetMyTickets)
		tickets.GET("/:id", h.GetTicket)
		tickets.POST("/:id/messages", h.AddCustomerMessage)
	}

	// Guest ticket (no auth required)
	rg.POST("/support/tickets/guest", h.CreateGuestTicket)

	// KB articles (public, no auth)
	rg.GET("/support/articles", h.ListArticles)
	rg.GET("/support/articles/:slug", h.GetArticle)

	helpful := rg.Group("/support/articles/:id/helpful")
	helpful.Use(requireAuth, requireCustomer)
	helpful.POST("", h.MarkHelpful)
}

func (h *Handler) RegisterStaffRoutes(rg *gin.RouterGroup, requireAuth, requireStaff gin.HandlerFunc) {
	tickets := rg.Group("/support/tickets")
	tickets.Use(requireAuth, requireStaff)
	{
		tickets.GET("", h.ListTickets)
		tickets.GET("/:id", h.GetTicketStaff)
		tickets.PATCH("/:id", h.UpdateTicket)
		tickets.POST("/:id/messages", h.AddStaffMessage)
		tickets.PATCH("/:id/assign", h.AssignTicket)
	}

	articles := rg.Group("/support/articles")
	articles.Use(requireAuth, requireStaff)
	{
		articles.POST("", h.CreateArticle)
		articles.PATCH("/:id", h.UpdateArticle)
		articles.DELETE("/:id", h.DeleteArticle)
	}
}


func (h *Handler) CreateTicket(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	var req CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ticket, err := h.svc.CreateTicket(c.Request.Context(), sid, cid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, ticket)
}

func (h *Handler) CreateGuestTicket(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req GuestTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ticket, err := h.svc.CreateGuestTicket(c.Request.Context(), sid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, ticket)
}

func (h *Handler) GetMyTickets(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit := int32(20)
	offset := int32((page - 1) * 20)
	tickets, err := h.svc.GetMyTickets(c.Request.Context(), sid, cid, limit, offset)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, tickets)
}

func (h *Handler) GetTicket(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid ticket id")
		return
	}
	ticket, msgs, err := h.svc.GetTicket(c.Request.Context(), sid, ticketID, cid, false)
	if err != nil {
		h.handleSvcError(c, err)
		return
	}
	response.OK(c, gin.H{"ticket": ticket, "messages": msgs})
}

func (h *Handler) AddCustomerMessage(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	cid, ok := ginutil.CustomerUUID(c)
	if !ok {
		return
	}
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid ticket id")
		return
	}
	var req AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = sid
	msg, err := h.svc.AddMessage(c.Request.Context(), sid, ticketID, cid, "customer", req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, msg)
}


func (h *Handler) ListArticles(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	articles, err := h.svc.ListArticles(c.Request.Context(), sid, category, true, 20, int32((page-1)*20))
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, articles)
}

func (h *Handler) GetArticle(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	art, err := h.svc.GetArticle(c.Request.Context(), sid, c.Param("slug"))
	if err != nil {
		if errors.Is(err, ErrArticleNotFound) {
			response.NotFound(c, "article not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, art)
}

func (h *Handler) MarkHelpful(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id")
		return
	}
	if err := h.svc.MarkHelpful(c.Request.Context(), articleID); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}


func (h *Handler) ListTickets(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	filter := TicketFilter{
		Status:   c.Query("status"),
		Priority: c.Query("priority"),
		Limit:    20,
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page > 1 {
		filter.Offset = int32((page - 1) * 20)
	}
	tickets, err := h.svc.ListTickets(c.Request.Context(), sid, filter)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OK(c, tickets)
}

func (h *Handler) GetTicketStaff(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid ticket id")
		return
	}
	ticket, msgs, err := h.svc.GetTicket(c.Request.Context(), sid, ticketID, uuid.Nil, true)
	if err != nil {
		h.handleSvcError(c, err)
		return
	}
	response.OK(c, gin.H{"ticket": ticket, "messages": msgs})
}

func (h *Handler) UpdateTicket(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid ticket id")
		return
	}
	var req UpdateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	ticket, err := h.svc.UpdateTicketStatus(c.Request.Context(), sid, ticketID, req)
	if err != nil {
		h.handleSvcError(c, err)
		return
	}
	response.OK(c, ticket)
}

func (h *Handler) AddStaffMessage(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	staffIDStr := c.GetString("staff_id")
	if staffIDStr == "" {
		response.Unauthorized(c, "staff authentication required")
		return
	}
	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		response.Unauthorized(c, "invalid staff token")
		return
	}
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid ticket id")
		return
	}
	var req AddMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = sid
	msg, err := h.svc.AddMessage(c.Request.Context(), sid, ticketID, staffID, "staff", req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, msg)
}

func (h *Handler) AssignTicket(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	ticketID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid ticket id")
		return
	}
	var body struct {
		StaffID string `json:"staff_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	staffID, err := uuid.Parse(body.StaffID)
	if err != nil {
		response.BadRequest(c, "invalid staff_id")
		return
	}
	ticket, err := h.svc.AssignTicket(c.Request.Context(), sid, ticketID, staffID)
	if err != nil {
		h.handleSvcError(c, err)
		return
	}
	response.OK(c, ticket)
}


func (h *Handler) CreateArticle(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	var req CreateKBArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	art, err := h.svc.CreateArticle(c.Request.Context(), sid, req)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.Created(c, art)
}

func (h *Handler) UpdateArticle(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id")
		return
	}
	var req UpdateKBArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	art, err := h.svc.UpdateArticle(c.Request.Context(), sid, articleID, req)
	if err != nil {
		if errors.Is(err, ErrArticleNotFound) {
			response.NotFound(c, "article not found")
			return
		}
		response.InternalError(c)
		return
	}
	response.OK(c, art)
}

func (h *Handler) DeleteArticle(c *gin.Context) {
	sid, ok := ginutil.ShopUUID(c)
	if !ok {
		return
	}
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid article id")
		return
	}
	if err := h.svc.DeleteArticle(c.Request.Context(), sid, articleID); err != nil {
		response.InternalError(c)
		return
	}
	response.NoContent(c)
}


func (h *Handler) handleSvcError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrTicketNotFound):
		response.NotFound(c, "ticket not found")
	case errors.Is(err, ErrForbidden):
		response.Forbidden(c, "access denied")
	default:
		response.InternalError(c)
	}
}
