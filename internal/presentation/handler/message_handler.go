package handler

import (
	"net/http"
	"strconv"

	"github.com/eneskaya/insider-messaging/internal/application/dto"
	"github.com/eneskaya/insider-messaging/internal/application/service"
	apperrors "github.com/eneskaya/insider-messaging/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MessageHandler struct {
	messageService service.MessageService
}

func NewMessageHandler(messageService service.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// GetSentMessages godoc
// @Summary Get list of sent messages
// @Description Retrieve a paginated list of successfully sent messages
// @Tags messages
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.MessageListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/messages/sent [get]
func (h *MessageHandler) GetSentMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.messageService.GetSentMessages(c.Request.Context(), page, pageSize)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetMessage godoc
// @Summary Get message by ID
// @Description Retrieve detailed information about a specific message
// @Tags messages
// @Accept json
// @Produce json
// @Param id path string true "Message ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/messages/{id} [get]
func (h *MessageHandler) GetMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid message ID format",
		})
		return
	}

	result, err := h.messageService.GetMessage(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetStats godoc
// @Summary Get message statistics
// @Description Retrieve statistics about messages (total, pending, sent, failed)
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {object} dto.MessageStatsResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/messages/stats [get]
func (h *MessageHandler) GetStats(c *gin.Context) {
	stats, err := h.messageService.GetStats(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, stats)
}

// CreateMessage godoc
// @Summary Create a new message
// @Description Create a new message to be sent
// @Tags messages
// @Accept json
// @Produce json
// @Param message body dto.CreateMessageRequest true "Message details"
// @Success 201 {object} dto.MessageResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/messages [post]
func (h *MessageHandler) CreateMessage(c *gin.Context) {
	var req dto.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	result, err := h.messageService.CreateMessage(c.Request.Context(), &req)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}
