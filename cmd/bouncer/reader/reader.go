package reader

import (
	"bufio"
	"errors"
	"os"
)

func ReadFile(job chan<- *bufio.Reader, fileName string) (err error) {
	file, err := os.Open(fileName)
	if err != nil {
		return err
		//return errors.New("no such file")
		//os.Exit(0)
	}
	defer func() {
		if err := file.Close(); err != nil {
			return
			//os.Exit(1)
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
