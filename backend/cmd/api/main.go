package main

import (
	"fmt"
	"os"

	"disk.simulator.com/m/v2/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Configurar CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.POST("/command", handlers.HandleCommand)
	r.POST("/login", handlers.HandleLogin)
	r.POST("/logout", handlers.HandleLogout)
	r.GET("/disks", handlers.HandleDisk)                      // Ruta para listar discos
	r.GET("/disks/partitions", handlers.HandleDiskPartitions) // Modificado para usar query param
	r.GET("/directory", handlers.HandleDirectoryLs)           // Nueva ruta para listar directorios en JSON
	r.POST("/read-file", handlers.HandleReadFile)             // Nueva ruta para leer contenido de archivos
	r.GET("/journaling", func(c *gin.Context) {               // Nueva ruta para obtener el journaling
		handlers.GetJournaling(c.Writer, c.Request)
	})

	filePath := "/discos/NAME.txt"
	err := os.WriteFile(filePath, []byte("Jorge"), 0644)
	if err != nil {
		fmt.Println("Error al crear NAME.txt:", err)
	}

	r.Run() // Por defecto escucha en :8080
}
