package reader

import (
	"errors"
	"fmt"
	"io"
	"net/mail"
	"os"
)

var (
	errInput = errors.New("no input specified")
	errFile  = errors.New("error reading file")
)

// ReceiveInput
type ReceiveInput struct {
	doneChannel chan bool
	JobChannel  chan *mail.Message
	fileNames   []string
}

// New creates a new ReceiveInput
func New(mailChan chan *mail.Message, fileNames []string) *ReceiveInput {
	return &ReceiveInput{
		doneChannel: make(chan bool),
		JobChannel:  mailChan,
		fileNames:   fileNames,
	}
}

// ScanInput reads input from stdin or files and passes it to the job channel
func (r *ReceiveInput) ScanInput() error {
	if len(r.fileNames) == 0 {
		if err := r.readStdin(); err != nil {
			return errInput
		}
		return nil
	}

	for _, fileName := range r.fileNames {
		err := r.readFile(fileName)
		if err != nil {
			fmt.Printf("Error reading file %s: %s", fileName, err)
			return errFile
		}
	}

	return nil
}

// readStdin reads eml file from stdin till EOF and passes it to the job channel
func (r *ReceiveInput) readStdin() error {
	// check if stdin is a pipe
	data, err := os.Stdin.Stat()
	if err != nil {
		return err
	}

	if (data.Mode() & os.ModeNamedPipe) == 0 {
		return errInput
	}

	return r.doJob(os.Stdin)
}

// readFile reads eml file and passes it to the job channel
func (r *ReceiveInput) readFile(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}

	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	return r.doJob(file)
}

// doJob reads eml from file or STDIN and passes it to the job channel
func (r *ReceiveInput) doJob(input io.Reader) error {
	m, err := mail.ReadMessage(input)
	if err != nil {
		return err
	}

	// send message to job channel
	r.JobChannel <- m

	// wait for job to finish
	<-r.doneChannel

	return nil
}

// printUsage prints usage information
func (r *ReceiveInput) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("bouncer -c config.yaml file.eml ...")
	fmt.Println("or")
	fmt.Println("cat file.eml | bouncer -c config.yaml")
	fmt.Println("or")
	fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
	os.Exit(0)
}

//// ScanInput reads input from stdin or files and passes it to the job channel
// func ScanInput(job chan<- *mail.Message, fileNames []string) {
//	if len(fileNames) == 0 {
//		if err := readStdin(job); err != nil {
//			fmt.Println("Usage:")
//			fmt.Println("bouncer -c config.yaml file.eml")
//			fmt.Println("or")
//			fmt.Println("cat file.eml | bouncer -c config.yaml")
//			fmt.Println("or")
//			fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
//			os.Exit(0)
//		}
//	}
//
//	for _, fileName := range fileNames {
//		readFile(job, fileName)
//	}
// }
//
// func readFile(job chan<- *mail.Message, fileName string) {
//	file, err := os.Open(fileName)
//	if err != nil {
//		return
//	}
//
//	defer func() {
//		if err := file.Close(); err != nil {
//			return
//		}
//	}()
//
//	m, err := mail.ReadMessage(file)
//	if err != nil {
//		return
//	}
//
//	job <- m
// }
//
// func readStdin(job chan<- *mail.Message) (err error) {
//	inputData, err := os.Stdin.Stat()
//	if err != nil {
//		panic(err)
//	}
//
//	if (inputData.Mode() & os.ModeNamedPipe) == 0 {
//		return errInput
//	}
//
//	m, err := mail.ReadMessage(os.Stdin)
//	if err != nil {
//		return err
//	}
//
//	job <- m
//
//	return nil
// }
