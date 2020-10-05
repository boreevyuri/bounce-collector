package reader

import (
	"bufio"
	"errors"
	"os"
)

func ReadFiles(job chan<- *bufio.Reader, fileNames []string) (err error) {
	for _, fileName := range fileNames {
		err = ReadFile(job, fileName)
		if err != nil {
			return
		}
	}
}

func ReadFile(job chan<- *bufio.Reader, fileName string) (err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			return
		}
	}()

	job <- bufio.NewReader(file)
	close(job)
	return nil
}

func ReadStdin(job chan<- *bufio.Reader) (err error) {
	inputData, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (inputData.Mode() & os.ModeNamedPipe) != 0 {
		//terminate()
		return errors.New("no input specified")
	}

	job <- bufio.NewReader(os.Stdin)
	close(job)
	return nil
}

//func terminate() {
//	fmt.Println("Usage:")
//	fmt.Println("bouncer -c config.yaml file.eml")
//	fmt.Println("or")
//	fmt.Println("cat file.eml | bouncer -c config.yaml")
//	fmt.Println("or")
//	fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
//	//os.Exit(0)
//}
