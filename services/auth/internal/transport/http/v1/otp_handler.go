package v1

import (
	"context"
	"net/http"
	"time"

	usecase "github.com/PavelLymin/diploma/services/auth/internal/application/use_case"
	autherrors "github.com/PavelLymin/diploma/services/auth/internal/transport/http"
	"github.com/PavelLymin/diploma/services/auth/internal/transport/http/v1/request"
	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
)

type OTPHandler struct {
	OTPUseCase     *usecase.UseCase
	requestTimeout time.Duration
}

func NewOTPHandler(
	uc *usecase.UseCase,
	requestTimeout time.Duration,
) *OTPHandler {
	return &OTPHandler{
		OTPUseCase:     uc,
		requestTimeout: requestTimeout,
	}
}

func (h *OTPHandler) CreateChallenge(c *gin.Context) {
	var req request.CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		autherrors.WriteError(c, err)
		return
	}

	ctx, cancel := h.requestContext(c)
	defer cancel()

	result, err := h.OTPUseCase.CreateChallenges(ctx, req)
	if err != nil {
		autherrors.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *OTPHandler) ResendChallenge(c *gin.Context) {
	var req request.ResendChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		autherrors.WriteError(c, err)
		return
	}

	ctx, cancel := h.requestContext(c)
	defer cancel()

	result, err := h.OTPUseCase.ResendChallenge(ctx, req)
	if err != nil {
		autherrors.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *OTPHandler) VerifyAndConsume(c *gin.Context) {
	var req request.VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		autherrors.WriteError(c, err)
		return
	}
	challangeID, err := uuid.Parse(c.Param("challenge_id"))
	if err != nil {
		autherrors.WriteError(c, err)
		return
	}

	ctx, cancel := h.requestContext(c)
	defer cancel()

	result, err := h.OTPUseCase.VerifyAndConsume(ctx, challangeID, req)
	if err != nil {
		autherrors.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *OTPHandler) requestContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), h.requestTimeout)
}
