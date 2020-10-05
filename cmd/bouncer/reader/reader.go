package reader

import (
	"errors"
	"net/mail"
	"os"
)

var (
	errInput = errors.New("no input specified")
)

func ReadFile(job chan<- *mail.Message, fileName string) (err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	m, err := mail.ReadMessage(file)
	if err != nil {
		return err
	}

	job <- m

	return nil
}

func ReadStdin(job chan<- *mail.Message) (err error) {
	inputData, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (inputData.Mode() & os.ModeNamedPipe) != 0 {
		//terminate()
		return errInput
	}

	m, err := mail.ReadMessage(os.Stdin)
	if err != nil {
		return err
	}

	job <- m

	return nil
}
