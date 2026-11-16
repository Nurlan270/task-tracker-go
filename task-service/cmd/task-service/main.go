package main

import "github.com/Nurlan270/task-tracker-go/task-service/internal/app"

func main() {
	application := app.New()

	application.Run()
}
