package main

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"bounce-collector/cmd/bouncer/reader"
	"bounce-collector/cmd/bouncer/writer"
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"net/mail"
	"os"
	"strings"
)

const (
	defaultConfigFile        = "/etc/bouncer.yaml"
	success           int    = 0
	runError          int    = 1
	failRedis         int    = 12
	pass              string = "Pass"
	decline           string = "Decline"
)

func main() {
	var (
		confFile  string
		checkAddr string
		conf      config.Conf
		//config    conf
	)

	flag.StringVar(&confFile, "c", defaultConfigFile, "configuration file")
	flag.StringVar(&checkAddr, "r", "", "email address to check existence")
	flag.Parse()

	//read config file
	conf.GetConf(confFile)

	//open connection to redis
	redis := writer.New(conf.Redis)

	//we have rcpt addr -> check and exit
	if len(checkAddr) != 0 {
		fmt.Printf("%s", checkMail(redis, checkAddr))
		os.Exit(success)
	}

	//we have no rcpt to check. Let mortal kombat begin...
	//fileName := flag.Arg(0)
	fileNames := flag.Args()

	//messageChan := make(chan *mail.Message)
	messageChan := make(chan *bufio.Reader)
	mailChan := make(chan *mail.Message)

	if len(fileNames) == 0 {
		if err := reader.ReadStdin(messageChan); err != nil {
			terminate()
		}
	} else {
		if err := reader.ReadFiles(messageChan, fileNames); err != nil {
			terminate()
		}
	}

	//
	//for _, fileName := range fileNames {
	//	processNewMail(redis, fileName)
	//}

	//processNewMail(redis, fileName)

	os.Exit(success)
}

func parseMail(m chan *mail.Message, r chan *bufio.Reader) {
	for message := range r {
		n, err := mail.ReadMessage(message)
		if err != nil {
			os.Exit(runError)
		}
		m <- n
	}
}

func checkMail(redis writer.ProcessRedis, emailAddr string) string {
	if redis.Find(emailAddr) {
		return pass
	}

	return decline
}

func processNewMail(redis writer.ProcessRedis, fileName string) {
	var (
		messageInfo analyzer.RecordInfo
		record      writer.Record
	)

	m := readInput(fileName)
	rcpt := strings.ToLower(m.Header.Get("X-Failed-Recipients"))
	body, _ := ioutil.ReadAll(m.Body)
	res := analyzer.Analyze(body)

	messageInfo = analyzer.RecordInfo{
		Domain:     getDomainFromAddress(rcpt),
		SMTPCode:   res.SMTPCode,
		SMTPStatus: res.SMTPStatus,
		Reason:     res.Reason,
		Date:       m.Header.Get("Date"),
		Reporter:   parseFrom(m.Header.Get("From")),
	}

	record = writer.Record{
		Rcpt: rcpt,
		TTL:  analyzer.SetTTL(messageInfo),
		Info: messageInfo,
	}

	if !redis.Insert(record) {
		os.Exit(failRedis)
	}
}

func readInput(fileName string) (m *mail.Message) {
	var reader *bufio.Reader

	inputData, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (inputData.Mode() & os.ModeNamedPipe) == 0 {
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Println("Usage:")
			fmt.Println("bouncer -c config.yaml file.eml")
			fmt.Println("or")
			fmt.Println("cat file.eml | bouncer -c config.yaml")
			fmt.Println("or")
			fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
			os.Exit(runError)
		}

		defer func() {
			if err := file.Close(); err != nil {
				os.Exit(runError)
			}
		}()

		reader = bufio.NewReader(file)
	} else {
		reader = bufio.NewReader(os.Stdin)
	}

	m, err = mail.ReadMessage(reader)

	if err != nil {
		os.Exit(runError)
	}

	return m
}

func getDomainFromAddress(addr string) string {
	a := strings.Split(addr, "@")
	if len(a) > 1 {
		return a[1]
	}

	return "unknown.tld"
}

func parseFrom(s string) string {
	e, err := mail.ParseAddress(s)
	if err != nil {
		return "unknown@unknown.tld"
	}

	return e.Address
}

func terminate() {
	fmt.Println("Usage:")
	fmt.Println("bouncer -c config.yaml file.eml")
	fmt.Println("or")
	fmt.Println("cat file.eml | bouncer -c config.yaml")
	fmt.Println("or")
	fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
	os.Exit(0)
}
