package console_executor

import (
	"log"
	"os/exec"
)

type ConsoleExecutorInterface interface {
	Execute(command string) (string, error)
}
type ConsoleExecutor struct{}

func (c *ConsoleExecutor) Execute(command string) (string, error) {
	log.Printf("Executing command: %s", command)
	stdout, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		log.Fatalf(err.Error())
	}

	return string(stdout), nil
}
