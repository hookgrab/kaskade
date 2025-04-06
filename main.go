package main

import (
	"github.com/gofiber/fiber/v2"

	"hg.atrin.dev/kaskade/routers"
	handlers "hg.atrin.dev/kaskade/handlers/socket"
)




func main() {
	workers := make(chan struct{}, 100)
	routers.SetWorkers(workers)

	go handlers.MessageLoop()

	app := fiber.New()

	routers.SetupRouters(app)

	app.Listen(":4000")
}
