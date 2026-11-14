package main

import "github.com/Nurlan270/task-tracker-go/user-service/internal/app"

func main() {
	application := app.New()

	application.Run()
}
