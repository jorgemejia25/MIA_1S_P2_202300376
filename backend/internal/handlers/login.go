package handlers

import (
	"fmt"

	"disk.simulator.com/m/v2/internal/disk/operations/auth"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Partition string `json:"partition"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}

func HandleLogin(c *gin.Context) {
	var req LoginRequest

	// Hacer bind del JSON al struct
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, LoginResponse{
			Success: false,
			Msg:     "Invalid request",
		})
		return
	}

	// Aquí puedes agregar la lógica para manejar el inicio de sesión
	user := req.Username
	password := req.Password
	partition := req.Partition

	fmt.Println("Login request:", user, password, partition)

	err := auth.Login(user, password, partition)

	if err != nil {
		c.JSON(400, LoginResponse{
			Success: false,
			Msg:     err.Error(),
		})
		return
	}

	c.JSON(200, LoginResponse{
		Success: true,
		Msg:     "Login successful",
	})

}
