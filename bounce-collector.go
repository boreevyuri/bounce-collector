package main

import (
	"bufio"
	"fmt"
	"github.com/boreevyuri/bounce-collector/analyzer"
	"io/ioutil"
	"net/mail"
	"os"
)

func main() {
	data, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if data.Mode()&os.ModeNamedPipe == 0 {
		fmt.Println("Use it with pipe")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	m, err := mail.ReadMessage(reader)
	if err != nil {
		panic(err)
	}

	// Забираем заголовки
	//headers := m.Header
	body, err := ioutil.ReadAll(m.Body)

	res := analyzer.Analyze(body)
	fmt.Printf("%v", res)
	//var mailMessage []string
	//for
	//	input, _, err := reader.Read()
	//	if err != nil && err == io.EOF {
	//		break
	//	}
	//	mailMessage = append(mailMessage, input)
	//}
}
