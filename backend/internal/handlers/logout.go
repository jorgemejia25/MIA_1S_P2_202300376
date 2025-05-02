package handlers

import (
	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	"github.com/gin-gonic/gin"
)

type LogoutResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

func HandleLogout(c *gin.Context) {
	auth.Logout()

	c.JSON(200, LogoutResponse{
		Success: true,
		Msg:     "Logout successful",
	})
}
