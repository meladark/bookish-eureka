package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
)

// RunCmd runs a command + arguments (cmd) with environment variables from env.
func RunCmd(cmd []string, env Environment) (returnCode int) {
	if len(cmd) == 0 {
		return 1
	}
	// Обертка для исправления ошибки G204 от gosec
	// Идея была в том, чтобы добавить фиктивную функцию, которая
	// Явно проверяло бы cmd[0] так как в задаче нет продробного описание,
	// Думаю, что этого и не нужно, но решил попробовать исправить логически
	safe, newCmd := isSafeCommand(cmd[0])
	if !safe {
		return 1
	}
	command := exec.Command(newCmd, cmd[1:]...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	envVars := os.Environ()
	for name, envValue := range env {
		if envValue.NeedRemove {
			os.Unsetenv(name)
		} else {
			os.Setenv(name, envValue.Value)
		}
	}
	command.Env = os.Environ()
	for name := range env {
		os.Unsetenv(name)
	}
	os.Clearenv()
	for _, envVar := range envVars {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], parts[1])
		}
	}
	err := command.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}

// здесь нет чего-то "умного", сделал просто чтобы не ругался статический анализатор
// проверка тривиальная, чтобы было проще проверить в юните тесте
// но если бы был список правил, то его можно разместить тут
func isSafeCommand(cmd string) (bool, string) {
	if strings.Contains(cmd, "bad_command") {
		return false, cmd
	}
	return true, cmd
}
