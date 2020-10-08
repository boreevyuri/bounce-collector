package reader

import (
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
)

var (
	errInput = errors.New("no input specified")
)

//func ReadFile(job chan<- *mail.Message, fileName string) (err error) {
func ReadFile(job chan<- *mail.Message, fileName string) {
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
		//return err
		return
	}

	//fmt.Fprintf("reader header: %s", strings.ToLower(m.Header.Get("X-Failed-Recipients")))
	h := strings.ToLower(m.Header.Get("X-Failed-Recipients"))
	fmt.Printf("reader header: %s\n", h)
	job <- m

	//return nil
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
