package main

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"bounce-collector/cmd/bouncer/reader"
	"bounce-collector/cmd/bouncer/writer"
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
	pass              string = "Pass"
	decline           string = "Decline"
	failRedis         int    = 12
	//runError          int    = 1
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
	fileNames := flag.Args()

	mailChan := make(chan *mail.Message)
	done := make(chan bool)

	go processMail(done, redis, mailChan)

	if len(fileNames) == 0 {
		if err := reader.ReadStdin(mailChan); err != nil {
			terminate()
		}
	} else {
		for _, fileName := range fileNames {
			reader.ReadFile(mailChan, fileName)
			//err := reader.ReadFile(mailChan, fileName)
			//fmt.Println("readfile started")
			//if err != nil {
			//	fmt.Printf("read file error")
			//}
		}
		close(mailChan)
	}
	<-done

	os.Exit(success)
}

func processMail(done chan<- bool, redis writer.ProcessRedis, in <-chan *mail.Message) {
	for m := range in {
		rcpt := strings.ToLower(m.Header.Get("X-Failed-Recipients"))

		body, _ := ioutil.ReadAll(m.Body)
		res := analyzer.Analyze(body)

		messageInfo := analyzer.RecordInfo{
			Domain:     getDomainFromAddress(rcpt),
			Reason:     res.Reason,
			Reporter:   parseFrom(m.Header.Get("From")),
			SMTPCode:   res.SMTPCode,
			SMTPStatus: res.SMTPStatus,
			Date:       m.Header.Get("Date"),
		}

		record := writer.Record{
			Rcpt: rcpt,
			TTL:  analyzer.SetTTL(messageInfo),
			Info: messageInfo,
		}

		success := redis.Insert(record)
		if !success {
			os.Exit(failRedis)
		}
	}

	done <- true

	//if !redis.Insert(record) {
	//	fmt.Println("record inserted")
	//	os.Exit(failRedis)
	//}

}

func checkMail(redis writer.ProcessRedis, emailAddr string) string {
	if redis.Find(emailAddr) {
		return pass
	}

	return decline
}

//func processNewMail(redis writer.ProcessRedis, fileName string) {
//	var (
//		messageInfo analyzer.RecordInfo
//		record      writer.Record
//	)
//
//	m := readInput(fileName)
//	rcpt := strings.ToLower(m.Header.Get("X-Failed-Recipients"))
//	body, _ := ioutil.ReadAll(m.Body)
//	res := analyzer.Analyze(body)
//
//	messageInfo = analyzer.RecordInfo{
//		Domain:     getDomainFromAddress(rcpt),
//		SMTPCode:   res.SMTPCode,
//		SMTPStatus: res.SMTPStatus,
//		Reason:     res.Reason,
//		Date:       m.Header.Get("Date"),
//		Reporter:   parseFrom(m.Header.Get("From")),
//	}
//
//	record = writer.Record{
//		Rcpt: rcpt,
//		TTL:  analyzer.SetTTL(messageInfo),
//		Info: messageInfo,
//	}
//
//	if !redis.Insert(record) {
//		os.Exit(failRedis)
//	}
//}

//func readInput(fileName string) (m *mail.Message) {
//	var reader *bufio.Reader
//
//	inputData, err := os.Stdin.Stat()
//	if err != nil {
//		panic(err)
//	}
//
//	if (inputData.Mode() & os.ModeNamedPipe) == 0 {
//		file, err := os.Open(fileName)
//		if err != nil {
//			fmt.Println("Usage:")
//			fmt.Println("bouncer -c config.yaml file.eml")
//			fmt.Println("or")
//			fmt.Println("cat file.eml | bouncer -c config.yaml")
//			fmt.Println("or")
//			fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
//			os.Exit(runError)
//		}
//
//		defer func() {
//			if err := file.Close(); err != nil {
//				os.Exit(runError)
//			}
//		}()
//
//		reader = bufio.NewReader(file)
//	} else {
//		reader = bufio.NewReader(os.Stdin)
//	}
//
//	m, err = mail.ReadMessage(reader)
//
//	if err != nil {
//		os.Exit(runError)
//	}
//
//	return m
//}

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
