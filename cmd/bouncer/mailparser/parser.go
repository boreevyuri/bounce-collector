package mailparser

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/database"
	"bufio"
	"log"
	"net/mail"
	"os"
)

type mailParser struct {
	input    chan []string
	analyzer *analyzer.MailAnalyzer
	db       database.DB
	Done     chan bool
}

// New creates new mailParser
func New(db database.DB) *mailParser {
	mp := new(mailParser)
	mp.db = db
	mp.input = make(chan []string)
	mp.Done = make(chan bool)
	mp.analyzer = analyzer.New()
	return mp
}

// ProcessMails - process mails from files or STDIN
func (mp *mailParser) ProcessMails(fileNames []string) {
	if len(fileNames) == 0 {
		// if no files, read from STDIN
		fileNames = append(fileNames, "")
	}
	go func() {
		for _, fileName := range fileNames {
			// read file
			data, err := mp.readFile(fileName)
			if err != nil {
				log.Println("unable to read file:", err)
				return
			}
			// TODO: parse file
			// TODO: send result to db
		}
		mp.Done <- true
	}()
}

// readFile - read file from disk or STDIN if fileName is empty
// returns *mail.Message or error
func (mp *mailParser) readFile(fileName string) (*mail.Message, error) {
	// create reader
	var reader *bufio.Reader

	// getting stat from stdin
	stdinData, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	// if stdin is not empty, use it
	if (stdinData.Mode() & os.ModeNamedPipe) != 0 {
		reader = bufio.NewReader(os.Stdin)
	} else {
		// otherwise read file
		file, err := os.Open(fileName)
		if err != nil {
			return nil, err
		}
		reader = bufio.NewReader(file)
	}

	// read message
	m, err := mail.ReadMessage(reader)
	if err != nil {
		return nil, err
	}

	return m, nil
}
