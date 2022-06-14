package cli

import (
	"fmt"
)

func Run() {
	InitAppContext()
	fmt.Println("Welcome to DeroSphere. Type 'help' for a list of commands")
	Context.Run()
}
