package v1

import (
	"net/http"

	usecase "github.com/PavelLymin/diploma/services/auth/internal/application/use_case"
	autherrors "github.com/PavelLymin/diploma/services/auth/internal/transport/http"
	"github.com/PavelLymin/diploma/services/auth/internal/transport/http/v1/request"
	"github.com/gin-gonic/gin"

	"github.com/google/uuid"
)

type OTPHandler struct {
	OTPUseCase usecase.UseCase
}

func New(OTPUseCase usecase.UseCase) *OTPHandler {
	return &OTPHandler{
		OTPUseCase: OTPUseCase,
	}
}

func (h *OTPHandler) CreateChallenge(c *gin.Context) {
	var req request.CreateChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		autherrors.WriteError(c, err)
		return
	}
	result, err := h.OTPUseCase.CreateChallenges(c, req)
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
	result, err := h.OTPUseCase.ResendChallenge(c, req)
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

	result, err := h.OTPUseCase.VerifyAndConsume(c, challangeID, req)
	if err != nil {
		autherrors.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
