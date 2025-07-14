package returns

import (
	"errors"
	"strconv"

	"backend-api/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, requireAuth gin.HandlerFunc, requireStaff gin.HandlerFunc) {
	ret := rg.Group("/returns", requireAuth, requireStaff)
	{
		ret.GET("", h.ListReturns)
		ret.POST("", h.CreateReturn)
		ret.GET("/:id", h.GetReturn)
		ret.PATCH("/:id/status", h.UpdateStatus)
		ret.POST("/:id/refund", h.ProcessRefund)
	}

	// Also accessible under /orders/:id/returns
	rg.GET("/orders/:id/returns", requireAuth, requireStaff, h.ListByOrder)
}


func (h *Handler) CreateReturn(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req CreateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ret, err := h.svc.CreateReturn(c.Request.Context(), shopID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.Created(c, ret)
}

func (h *Handler) GetReturn(c *gin.Context) {
	shopID := c.GetString("shop_id")

	ret, err := h.svc.GetReturn(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, ret)
}

func (h *Handler) ListReturns(c *gin.Context) {
	shopID := c.GetString("shop_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	rets, total, err := h.svc.ListReturns(c.Request.Context(), shopID, page, perPage)
	if err != nil {
		response.InternalError(c)
		return
	}
	response.OKWithMeta(c, rets, response.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		TotalItems: total,
		TotalPages: response.CalcTotalPages(total, int64(perPage)),
	})
}

func (h *Handler) ListByOrder(c *gin.Context) {
	shopID := c.GetString("shop_id")

	rets, err := h.svc.ListByOrder(c.Request.Context(), shopID, c.Param("id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, rets)
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req UpdateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	ret, err := h.svc.UpdateStatus(c.Request.Context(), shopID, c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, ret)
}

func (h *Handler) ProcessRefund(c *gin.Context) {
	shopID := c.GetString("shop_id")

	var req ProcessRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Body is optional only when intending a zero-amount refund.
		// A malformed body is still an error.
		if err.Error() != "EOF" {
			response.BadRequest(c, err.Error())
			return
		}
	}

	ret, err := h.svc.ProcessRefund(c.Request.Context(), shopID, c.Param("id"), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	response.OK(c, ret)
}


func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidStatus):
		response.BadRequest(c, err.Error())
	case errors.Is(err, ErrAlreadyRefunded),
		errors.Is(err, ErrBadTransition):
		response.UnprocessableEntity(c, err.Error())
	default:
		response.HandleError(c, err)
	}
}
