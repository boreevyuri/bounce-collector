package main

import (
	"bufio"
	"fmt"
	"github.com/boreevyuri/bounce-collector/analyzer"
	"io/ioutil"
	"net/mail"
	"os"
	"strings"
)

func main() {
	var m *mail.Message

	data, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (data.Mode() & os.ModeNamedPipe) == 0 {
		file, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("Use with pipe or put file as first argument")
			panic(err)
		}

		reader := strings.NewReader(string(file))
		m, err = mail.ReadMessage(reader)

		if err != nil {
			panic(err)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		m, err = mail.ReadMessage(reader)

		if err != nil {
			panic(err)
		}
	}

	// Забираем заголовки
	//headers := m.Header
	body, _ := ioutil.ReadAll(m.Body)

	res := analyzer.Analyze(body)
	fmt.Printf("%v", res)
}
