package app

import (
	"github.com/Nurlan270/task-tracker-go/common/config"

	appcfg "github.com/Nurlan270/task-tracker-go/auth-service/internal/config"
)

type App struct {
	cfg *appcfg.Config
	di  diContainer
}

func New() *App {
	app := &App{
		cfg: config.MustLoad[appcfg.Config](),
	}

	app.di = newDIContainer(app.cfg)

	app.initDeps()

	return app
}

func (a *App) initDeps() {
	deps := []func(){
		//
	}

	for _, init := range deps {
		init()
	}
}
