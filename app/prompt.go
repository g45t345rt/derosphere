package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chzyer/readline"
)

func Prompt(prompt string, defaultValue string) (string, error) {
	i := Context.readlineInstance
	Context.StopPromptRefresh = true

	if defaultValue != "" {
		i.SetPrompt(fmt.Sprintf("%s [%s]: ", prompt, defaultValue))
	} else {
		i.SetPrompt(fmt.Sprintf("%s: ", prompt))
	}

	line, err := i.Readline()
	Context.StopPromptRefresh = false
	if err != nil {
		return "", err
	}

	if line == "" {
		line = defaultValue
	}

	return line, nil
}

func PromptInt(prompt string, defaultValue int64) (int64, error) {
	valueString, err := Prompt(prompt, fmt.Sprintf("%d", defaultValue))
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(valueString, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

func PromptYesNo(prompt string, defaultAnswer bool) (bool, error) {
	defaultString := "y"
	if !defaultAnswer {
		defaultString = "n"
	}

	value, err := Prompt(fmt.Sprintf("%s (y/n)", prompt), defaultString)
	if err != nil {
		return false, err
	}

	if value == "y" {
		return true, nil
	} else {
		return false, nil
	}
}

func PromptChoose(prompt string, choices []string, defaultValue string) (string, error) {
	i := Context.readlineInstance
	Context.StopPromptRefresh = true

prompt:
	if defaultValue != "" {
		i.SetPrompt(fmt.Sprintf("%s (%s) [%s]: ", prompt, strings.Join(choices, "/"), defaultValue))
	} else {
		i.SetPrompt(fmt.Sprintf("%s (%s): ", prompt, strings.Join(choices, "/")))
	}

	line, err := i.Readline()
	if err != nil {
		Context.StopPromptRefresh = false
		return "", err
	}

	if line == "" {
		line = defaultValue
	}

	valid := false
	for _, v := range choices {
		if v == line {
			valid = true
		}
	}

	if !valid {
		fmt.Printf("# Enter %s or %s\n", strings.Join(choices[:len(choices)-1], ", "), choices[len(choices)-1])
		goto prompt
	}

	Context.StopPromptRefresh = false
	return line, nil
}

func PromptPassword(prompt string) (string, error) {
	i := Context.readlineInstance
	Context.StopPromptRefresh = true

	config := i.GenPasswordConfig()
	config.SetListener(func(line []rune, pos int, key rune) (newLine []rune, newPos int, ok bool) {
		i.SetPrompt(fmt.Sprintf("%s(%v): ", prompt, len(line)))
		i.Refresh()
		return nil, 0, false
	})

	line, err := i.ReadPasswordWithConfig(config)
	Context.StopPromptRefresh = false
	if err != nil {
		return "", err
	}

	return string(line), nil
}

func HandlePromptErr(err error) bool {
	if err != nil {
		if err == readline.ErrInterrupt {
			fmt.Println("Prompt cancelled")
		} else {
			fmt.Println(err)
		}

		return true
	}

	return false
}
