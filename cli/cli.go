package cli

import (
	"fmt"

	"github.com/g45t345rt/derosphere/app"
)

func Run() {
	app.InitAppContext(RootApp(), WalletApp())
	fmt.Println("Welcome to DeroSphere. Type 'help' for a list of commands")
	app.Context.Run()
}
