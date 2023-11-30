package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Body  string `json:"body"`
}

type Done struct {
	Start int `json:"start"`
	End   int `json:"end"` // non inclusive
}

func getValues(m map[int]Todo) []Todo {
	values := make([]Todo, 0, len(m))

	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func main() {
	fmt.Print("Hello world")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", // this is the ip and port for the client
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// map id to Todo
	todos := make(map[int]Todo)

	app.Get("/healthcheck", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}

		if err := c.BodyParser(todo); err != nil {
			return err
		}

		todo.ID = len(todos)

		todos[todo.ID] = *todo

		return c.JSON(getValues((todos)))

	})

	app.Patch("/api/todos/:id/done", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(401).SendString("Invalid id")
		}

		if todo, ok := todos[id]; ok {
			todo.Done = true
			todos[id] = todo
		}

		return c.JSON(getValues(todos))
	})

	// toggle Done
	app.Patch("/api/todos/:id/toggle", func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")

		if err != nil {
			return c.Status(401).SendString("Invalid id")
		}

		if todo, ok := todos[id]; ok {
			todo.Done = !todo.Done
			todos[id] = todo
		}

		return c.JSON(getValues(todos))
	})

	// makes multiple items done at once
	app.Patch("/api/todos/done", func(c *fiber.Ctx) error {

		done := &Done{}

		if err := c.BodyParser(done); err != nil {
			return err
		}

		for i := done.Start; i < done.End; i++ {
			if todo, ok := todos[i]; ok {
				todo.Done = true
				todos[i] = todo
			}
		}

		return c.JSON(getValues((todos)))
	})

	app.Get("/api/todos", func(c *fiber.Ctx) error {

		return c.JSON(getValues(todos))
	})

	log.Fatal(app.Listen(":4000")) // port 4000 for this server

}
