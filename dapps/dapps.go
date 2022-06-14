package dapps

import (
	"github.com/g45t345rt/derosphere/dapps/lotto"
	"github.com/urfave/cli/v2"
)

type DApp struct {
	Name        string
	Description string
	Version     string
	Commands    []*cli.Command
}

func FindDApp(name string) *DApp {
	for _, dapp := range GetDApps() {
		if dapp.Name == name {
			return &dapp
		}
	}

	return nil
}

func GetDApps() []DApp {
	return []DApp{
		{
			Name:        "nameservice",
			Description: "Register multiple names to receive DERO from others.",
			Version:     "0.0.1",
			Commands:    []*cli.Command{},
		},
		{
			Name:        "dapp-username",
			Description: "Register a single username used by other dApps.",
			Version:     "0.0.1",
			Commands:    []*cli.Command{},
		},
		{
			Name:        "lotto",
			Description: "Official custom lottery pools. Create your own type of lottery.",
			Version:     "0.0.1",
			Commands:    lotto.Commands(),
		},
	}
}
