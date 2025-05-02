package main

import (
	"disk.simulator.com/m/v2/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.POST("/command", handlers.HandleCommand)
	r.POST("/login", handlers.HandleLogin)
	r.POST("/logout", handlers.HandleLogout)
	r.GET("/disks", handlers.HandleDisk)                      // Ruta para listar discos
	r.GET("/disks/partitions", handlers.HandleDiskPartitions) // Modificado para usar query param

	r.Run() // Por defecto escucha en :8080
}
