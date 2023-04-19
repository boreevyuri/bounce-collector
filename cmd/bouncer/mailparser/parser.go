package mailparser

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/database"
	"bufio"
	"github.com/boreevyuri/go-email/email"
	"log"
	"os"
	"time"
)

type MailParser struct {
	input    chan []string
	analyzer *analyzer.MailAnalyzer
	db       database.DB
	Done     chan bool
}

// New creates new MailParser
func New(db database.DB) *MailParser {
	mp := new(MailParser)
	mp.db = db
	mp.input = make(chan []string)
	mp.Done = make(chan bool)
	mp.analyzer = analyzer.New()
	return mp
}

// ProcessMails - process mails from files or STDIN
func (mp *MailParser) ProcessMails(fileNames []string) {
	if len(fileNames) == 0 {
		// if no files, read from STDIN
		fileNames = append(fileNames, "")
	}

	go func() {
		for _, fileName := range fileNames {
			// read file
			data, err := mp.readEmail(fileName)
			if err != nil {
				log.Println("unable to read file:", err)
				return
			}

			// parse file
			result, err := mp.analyzer.Do(data)
			if err != nil {
				log.Println("unable to parse file:", err)
				return
			}
			log.Printf("result: %+v", result)

			// write to db
			mp.write(result)
		}
		mp.Done <- true
	}()
}

// readEmail - read file from disk or STDIN if fileName is empty
// returns *mail.Message or error
func (mp *MailParser) readEmail(fileName string) (*email.Message, error) {
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
	m, err := email.ParseMessage(reader)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// write - writes data to db
func (mp *MailParser) write(data analyzer.RecordInfo) {
	m, err := data.ToJSON()
	if err != nil {
		log.Println("unable to marshal record info:", err)
		return
	}

	if data.Rcpt == "" {
		log.Println("rcpt is empty. Skip writing to db")
		return
	}

	record := database.RecordPayload{
		Key:   data.Rcpt,
		Value: m,
		TTL:   time.Second * 600,
	}
	// send record to db
	if !mp.db.Insert(record) {
		log.Println("unable to insert data to db")
		return
	}
}
