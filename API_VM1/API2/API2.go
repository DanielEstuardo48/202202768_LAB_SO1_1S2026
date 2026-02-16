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
			"message":   "API2 is Ready",
			"timestamp": time.Now().Format(time.RFC3339),
			"VM":        "VM1",
			"carnet":    "202202768",
		})
	})

	// Llamada a API1
	app.Get("/api2/202202768/call-api1", func(c *fiber.Ctx) error {
		resp, err := http.Get("http://192.168.122.247:8881/health")

		if err != nil || resp.StatusCode != 200 {
			return c.Status(500).JSON(fiber.Map{
				"apiname":    "API1",
				"message":    "ERROR: The API1 located on the VM1 is not working",
				"connection": false,
				"carnet":     "202202768",
			})
		}

		return c.JSON(fiber.Map{
			"apiname":    "API1",
			"message":    "The API1 located on the VM1 is working",
			"connection": true,
			"carnet":     "202202768",
		})
	})

	// Llamada a API3
	app.Get("/api2/202202768/call-api3", func(c *fiber.Ctx) error {
		resp, err := http.Get("http://192.168.122.187:8883/health")

		if err != nil || resp.StatusCode != 200 {
			return c.Status(500).JSON(fiber.Map{
				"apiname":    "API3",
				"message":    "ERROR: The API3 located on the VM1 is not working",
				"connection": false,
				"carnet":     "202202768",
			})
		}

		return c.JSON(fiber.Map{
			"apiname":    "API3",
			"message":    "The API3 located on the VM1 is working",
			"connection": true,
			"carnet":     "202202768",
		})
	})
	// Inicia el servidor
	log.Fatal(app.Listen(":8882"))
}
