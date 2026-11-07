package main

import "github.com/Nurlan270/task-tracker-go/gateway/internal/app"

func main() {
	application := app.New()

	application.Run()
}
