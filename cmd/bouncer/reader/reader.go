package reader

import (
	"errors"
	"fmt"
	"net/mail"
	"os"
)

var (
	errInput = errors.New("no input specified")
)

func ReadInput(job chan<- *mail.Message, fileNames []string) {
	if len(fileNames) == 0 {
		if err := readStdin(job); err != nil {
			fmt.Println("Usage:")
			fmt.Println("bouncer -c config.yaml file.eml")
			fmt.Println("or")
			fmt.Println("cat file.eml | bouncer -c config.yaml")
			fmt.Println("or")
			fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
			os.Exit(0)
		}
	}

	for _, fileName := range fileNames {
		readFile(job, fileName)
	}
}

func readFile(job chan<- *mail.Message, fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		return
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	m, err := mail.ReadMessage(file)
	if err != nil {
		return
	}

	job <- m
}

func readStdin(job chan<- *mail.Message) (err error) {
	inputData, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (inputData.Mode() & os.ModeNamedPipe) == 0 {
		return errInput
	}

	m, err := mail.ReadMessage(os.Stdin)
	if err != nil {
		return err
	}

	job <- m

	return nil
}
