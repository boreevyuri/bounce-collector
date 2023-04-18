package main

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"bounce-collector/cmd/bouncer/database"
	"bounce-collector/cmd/bouncer/writer"
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"
)

const (
	// DefaultConfigFile sets default config file
	DefaultConfigFile     = "/etc/bouncer.conf"
	success           int = 0
	runError          int = 1
	failConfig        int = 13
	failRedis         int = 12
)

func main() {
	var (
		confFile  string
		checkAddr string
		// config    conf
	)

	flag.StringVar(&confFile, "c", DefaultConfigFile, "configuration file")
	flag.StringVar(&checkAddr, "r", "", "email address to check existence")
	flag.Parse()

	// read config file
	conf := new(config.Conf)
	err := conf.Parse(confFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(failConfig)
	}

	// create database connection
	db, err := database.NewDB(conf)
	if err != nil {
		fmt.Println(err)
		os.Exit(failConfig)
	}

	// if there is address to check, check it and exit
	if len(checkAddr) != 0 {
		switch db.Find(strings.ToLower(checkAddr)) {
		case true:
			// record found, let exim router pass it
			fmt.Println("Pass")
		case false:
			// record not found, let exim router decline it
			fmt.Println("Decline")
		}
		// close database connection
		db.Close()
		os.Exit(success)
	}

	// Check if we need to process mail from file or check address
	fileName := flag.Arg(0)

	// get mail from stdin or file and process it
	processMail(fileName, conf.Redis)

	os.Exit(success)
}

func processMail(fileName string, redisConfig config.RedisConfig) {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// create simple context
	ctx := context.Background()

	var (
		messageInfo analyzer.RecordInfo
		record      writer.Record
	)

	m := readInput(fileName)
	rcpt := strings.ToLower(m.Header.Get("X-Failed-Recipients"))
	body, _ := io.ReadAll(m.Body)
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

	err := writer.PutRecord(ctx, record, redisConfig)
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
			fmt.Println("bounce-collector -c config.yaml file.eml")
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
