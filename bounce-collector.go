package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boreevyuri/bounce-collector/analyzer"
	"github.com/boreevyuri/bounce-collector/writer"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/mail"
	"os"
	"strings"
)

const (
	ExampleConfig     = "./bounce-collector.conf"
	success       int = 0
	failRedis     int = 2
)

type conf struct {
	Redis writer.Config `yaml:"redis"`
}

func main() {
	var file string

	flag.StringVar(&file, "f", ExampleConfig, "configuration file")
	flag.Parse()

	var config conf

	config.getConf(file)

	var m *mail.Message

	data, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if (data.Mode() & os.ModeNamedPipe) == 0 {
		file, err := ioutil.ReadFile(os.Args[1])
		if err != nil {
			fmt.Println("Usage:")
			fmt.Println("bounce-collector file.eml")
			fmt.Println("or")
			fmt.Println("cat file.eml | bounce-collector")
			panic(err)
		}

		reader := strings.NewReader(string(file))
		m, err = mail.ReadMessage(reader)

		if err != nil {
			panic(err)
		}
	} else {
		reader := bufio.NewReader(os.Stdin)
		m, err = mail.ReadMessage(reader)

		if err != nil {
			panic(err)
		}
	}

	//Достаем реципиента из X-Failed-Recipients
	rcpt, domain := getFailedRcpt(m.Header)
	reporter := parseFrom(m.Header.Get("From"))
	date := m.Header.Get("Date")

	//Забираем тело на анализы
	body, _ := ioutil.ReadAll(m.Body)
	res := analyzer.Analyze(body)

	messageInfo := &analyzer.RecordInfo{
		Domain:     domain,
		Reason:     res.Reason,
		Reporter:   reporter,
		SMTPCode:   res.SMTPCode,
		SMTPStatus: res.SMTPStatus,
		Date:       date,
	}

	record := &writer.Record{
		Rcpt: rcpt,
		TTL:  analyzer.SetTTL(messageInfo),
		Info: marshalInfo(messageInfo),
	}

	err = writer.PutRecord(record, config.Redis)
	if err != nil {
		fmt.Printf("Collector error: %+v", err)
		os.Exit(failRedis)
	}

	os.Exit(success)
}

func marshalInfo(r *analyzer.RecordInfo) string {
	e, err := json.Marshal(r)
	if err != nil {
		return ""
	}

	return string(e)
}

func getFailedRcpt(h mail.Header) (addr string, domain string) {
	addr = strings.ToLower(h.Get("X-Failed-Recipients"))
	components := strings.Split(addr, "@")

	if len(components) == 2 {
		return addr, components[1]
	}

	return addr, "unknown"
}

func parseFrom(s string) string {
	e, err := mail.ParseAddress(s)
	if err != nil {
		return "unknown@unknown.domain"
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
			panic("no config file specified")
		}

		err = yaml.Unmarshal(yamlFile, c)

		if err != nil {
			fmt.Println(yamlFile)
			panic("Invalid config")
		}

		return c
	}

	return nil
}
