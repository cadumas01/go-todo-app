package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/lib/pq"
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

type Credentials struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

func getValues(m map[int]Todo) []Todo {
	values := make([]Todo, 0, len(m))

	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// returns a list of Todos after query from DB
func selectAll(db *sql.DB) []Todo {
	sql := fmt.Sprintf("SELECT * FROM %s.%s;", schemaName, tableName)
	rows, err := db.Query(sql)
	if err != nil {
		panic(err)
	}

	todos := []Todo{}

	for rows.Next() {
		var todo Todo
		err = rows.Scan(&todo.ID, &todo.Title, &todo.Done, &todo.Body)
		if err != nil {
			panic(err)
		}
		todos = append(todos, todo)
	}

	return todos
}

const schemaName = "golang_todo"
const tableName = "todos"

func main() {
	fmt.Print("Hello world")

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173", // this is the ip and port for the client
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// load credentials from db_credentials.json
	credsFile, _ := os.Open("db_credentials.json")

	credsBytes, _ := io.ReadAll(credsFile)

	var creds Credentials
	json.Unmarshal(credsBytes, &creds)

	// connect to sql database
	// https://www.calhoun.io/connecting-to-a-postgresql-database-with-gos-database-sql-package/
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		creds.Host, creds.Port, creds.User, creds.DBName) // could add in password into string later

	// validate connection arguments to db
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// connect db
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// initialize table if it doesn't exist
	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s(
			id		SERIAL PRIMARY KEY NOT NULL,
			title	TEXT,
			done	BOOLEAN,
			body	TEXT
		);
	`, schemaName, tableName) // formatted with schema and table name
	_, err = db.Exec(sql)

	if err != nil {
		fmt.Println("Failed to make db")
	}

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

		// ID is determined by number of id's in table
		// DO THIS

		todo.ID = len(todos)

		// remove this once sql works
		todos[todo.ID] = *todo

		// formatted insert to have schemaName and tableName but other values are
		// inserted in db.Exec
		sql := fmt.Sprintf(`
			INSERT INTO %s.%s (title, done, body)
			VALUES ($1, $2, $3);
		`, schemaName, tableName)
		_, err := db.Exec(sql, todo.Title, todo.Done, todo.Body)
		if err != nil {
			fmt.Println("failed to insert rows")
		}

		return c.JSON(selectAll(db))

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

		return c.JSON(selectAll(db))
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

		return c.JSON(selectAll(db))
	})

	app.Get("/api/todos", func(c *fiber.Ctx) error {

		return c.JSON(selectAll(db))
	})

	log.Fatal(app.Listen(":4000")) // port 4000 for this server

}
