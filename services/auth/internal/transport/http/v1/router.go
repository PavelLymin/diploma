package v1

import (
	"github.com/gin-gonic/gin"
)

func NewRouter(engine *gin.Engine, h *OTPHandler) {
	v1Group := engine.Group("/api/v1")
	{
		auth := v1Group.Group("/auth")
		{
			auth.POST("/challenges", h.CreateChallenge)
			auth.POST("/challenges/:challenge_id/verify", h.VerifyAndConsume)
			auth.POST("/challenges/:challenge_id/resend", h.ResendChallenge)
			// auth.POST("/refresh", h.Refresh)
			// auth.POST("/logout", h.Logout)
		}
	}

}
