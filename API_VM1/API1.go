package main

import (
	"log" // Para mensajes a la salida
	"time"

	"github.com/gofiber/fiber/v2"                   // Framework web para manejar solicitudes HTTP
	"github.com/gofiber/fiber/v2/middleware/logger" // Middleware para registrar solicitudes HTTP
)

func main() {
	// Crear nueva instancia de Fiber
	app := fiber.New()

	app.Use(logger.New()) // agregar middleware de loggin a la aplicacion

	// Definir ruta get en "/health"
	app.Get("/health", func(c *fiber.Ctx) error {

		// Respuesta con un json
		return c.JSON(map[string]string{
			"status":    "UP",
			"message":   "API1 is Ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"VM":        "VM1",
			"carnet":    "202202768",
		})
	})

	// Inicia el servidor
	log.Fatal(app.Listen(":8881"))
}
