package main

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
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

	fileName := flag.Arg(0)

	conf.GetConf(confFile)

	redis := writer.New(conf.Redis)

	//time.Sleep(time.Duration(2) * time.Second)

	if len(checkAddr) == 0 {
		processNewMail(redis, fileName)
	} else {
		fmt.Printf("%s", checkMail(redis, checkAddr))
	}

	os.Exit(success)
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
			fmt.Println("bounce-collector -c config.yaml file.eml")
			fmt.Println("or")
			fmt.Println("cat file.eml | bounce-collector -c config.yaml")
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
