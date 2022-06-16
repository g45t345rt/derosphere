package dapps

import (
	"github.com/g45t345rt/derosphere/dapps/dnd"
	"github.com/g45t345rt/derosphere/dapps/lotto"
	"github.com/g45t345rt/derosphere/dapps/nameservice"
	"github.com/g45t345rt/derosphere/dapps/seals"
	"github.com/g45t345rt/derosphere/dapps/username"
	"github.com/urfave/cli/v2"
)

func Find(name string) *cli.App {
	for _, app := range List() {
		if app.Name == name {
			return app
		}
	}

	return nil
}

func List() []*cli.App {
	return []*cli.App{
		nameservice.App(),
		username.App(),
		lotto.App(),
		dnd.App(),
		seals.App(),
	}
}
