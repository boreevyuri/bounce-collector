package main

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/writer"
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/mail"
	"os"
	"strings"
)

const (
	//DefaultConfigFile sets default config file
	DefaultConfigFile     = "/etc/bouncer.yaml"
	success           int = 0
	runError          int = 1
	failConfig        int = 13
	failRedis         int = 12
)

type conf struct {
	Redis writer.Config `yaml:"redis"`
}

func main() {
	var (
		confFile  string
		checkAddr string
		config    conf
	)

	flag.StringVar(&confFile, "c", DefaultConfigFile, "configuration file")
	flag.StringVar(&checkAddr, "r", "", "email address to check existence")
	flag.Parse()

	fileName := flag.Arg(0)

	config.getConf(confFile)

	if len(checkAddr) == 0 {
		processMail(fileName, config.Redis)
	} else {
		msg := checkMail(checkAddr, config.Redis)
		fmt.Println(msg)
	}

	os.Exit(success)
}

func checkMail(email string, redis writer.Config) string {
	if writer.IsPresent(email, redis) {
		return "Pass"
	}

	return "Decline"
}

func processMail(fileName string, redis writer.Config) {
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

	err := writer.PutRecord(record, redis)
	if err != nil {
		fmt.Printf("Collector error: %+v", err)
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
			fmt.Println("bounce-collector -f config.yaml file.eml")
			fmt.Println("or")
			fmt.Println("cat file.eml | bounce-collector -f config.yaml")
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

func isValidConfigFilename(filename string) bool {
	return len(filename) > 0
}

func (c *conf) getConf(filename string) *conf {
	if isValidConfigFilename(filename) {
		yamlFile, err := ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("no config file specified")
			os.Exit(failConfig)
		}

		err = yaml.Unmarshal(yamlFile, c)

		if err != nil {
			fmt.Println(yamlFile)
			os.Exit(failConfig)
		}

		return c
	}

	return nil
}
