package main

import "github.com/Nurlan270/task-tracker-go/auth-service/internal/app"

func main() {
	application := app.New()

	application.Run()
}
