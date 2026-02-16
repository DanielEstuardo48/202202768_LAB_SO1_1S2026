package main

import (
	"log"      // Para mensajes a la salida
	"net/http" // Para realizar solicitudes HTTP
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
			"message":   "API3 is Ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"VM":        "VM2",
			"carnet":    "202202768",
		})
	})

	// Llamada a API1
	app.Get("/api3/202202768/call-api1", func(c *fiber.Ctx) error {
		resp, err := http.Get("http://192.168.122.247:8881/health")

		if err != nil || resp.StatusCode != 200 {
			return c.Status(500).JSON(fiber.Map{
				"apiname":    "API1",
				"message":    "ERROR: The API1 located on the VM2 is not working",
				"connection": false,
				"carnet":     "202202768",
			})
		}

		return c.JSON(fiber.Map{
			"apiname":    "API1",
			"message":    "The API1 located on the VM2 is working",
			"connection": true,
			"carnet":     "202202768",
		})
	})

	// Llamada a API2
	app.Get("/api3/202202768/call-api2", func(c *fiber.Ctx) error {
		resp, err := http.Get("http://192.168.122.247:8882/health")

		if err != nil || resp.StatusCode != 200 {
			return c.Status(500).JSON(fiber.Map{
				"apiname":    "API2",
				"message":    "ERROR: The API2 located on the VM2 is not working",
				"connection": false,
				"carnet":     "202202768",
			})
		}

		return c.JSON(fiber.Map{
			"apiname":    "API2",
			"message":    "The API2 located on the VM2 is working",
			"connection": true,
			"carnet":     "202202768",
		})
	})

	// Inicia el servidor
	log.Fatal(app.Listen(":8883"))
}
