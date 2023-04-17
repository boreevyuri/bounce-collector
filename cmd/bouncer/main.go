package main

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"bounce-collector/cmd/bouncer/db"
	"bounce-collector/cmd/bouncer/reader"
	"flag"
	"fmt"
	"net/mail"
	"os"
	"strings"
)

const (
	// defaultConfigFile is the default path to the config file
	defaultConfigFile = "/etc/bouncer.yaml"
	// defaultRedisAddress is the default address of the redis server
	defaultRedisAddress = "127.0.0.1:6379"
	// defaultRedisPassword is the default password of the redis server
	defaultRedisPassword = ""
	// defaultRedisDB is the default database of the redis server
	defaultRedisDB = 0
	// success is the exit code for successful execution
	success int    = 0
	pass    string = "Pass"
	decline string = "Decline"
	// failRedis         int    = 12
	// runError          int    = 1
)

func main() {
	var (
		confFile  string
		checkAddr string
		conf      config.Conf
	)

	flag.StringVar(&confFile, "c", defaultConfigFile, "configuration file")
	flag.StringVar(&checkAddr, "r", "", "email address to check existence")
	flag.Parse()

	// read config file
	conf.GetConf(confFile)

	// open connection to db
	dbConnection := db.New(conf)

	// we have rcpt addr -> check and exit
	if len(checkAddr) != 0 {
		fmt.Printf("%s", checkMail(dbConnection, checkAddr))
		os.Exit(success)
	}

	// we have no rcpt to check. Let mortal kombat begin...
	fileNames := flag.Args()

	mailChan := make(chan *mail.Message)
	done := make(chan bool)

	// go processMail(done, dbConnection, mailChan)
	go processMail(done, mailChan)

	reader := reader.New(mailChan, fileNames)
	reader.ScanInput()

	//reader.ScanInput(mailChan, fileNames)

	close(mailChan)
	<-done

	os.Exit(success)
}

// func processMail(done chan<- bool, redis db.DBInterface, in <-chan *mail.Message) {
func processMail(done chan<- bool, in <-chan *mail.Message) {
	for m := range in {

		domain, err := analyzer.NewAnalyze(m)
		if err != nil {
			fmt.Printf("Error!!!")
		}
		fmt.Printf("Domain: %s", domain)
		//rcpt := strings.ToLower(m.Header.Get("X-Failed-Recipients"))

		//body, _ := ioutil.ReadAll(m.Body)
		//res := analyzer.Analyze(body)
		//
		//messageInfo := analyzer.RecordInfo{
		//	Domain:     getDomainFromAddress(rcpt),
		//	Reason:     res.Reason,
		//	Reporter:   parseFrom(m.Header.Get("From")),
		//	SMTPCode:   res.SMTPCode,
		//	SMTPStatus: res.SMTPStatus,
		//	Date:       m.Header.Get("Date"),
		//}
		//
		//record := writer.Record{
		//	Rcpt: rcpt,
		//	TTL:  analyzer.SetTTL(messageInfo),
		//	Info: messageInfo,
		//}
		//
		//success := redis.Insert(record)
		//if !success {
		//	os.Exit(failRedis)
		//}
	}

	done <- true

	//if !redis.Insert(record) {
	//	fmt.Println("record inserted")
	//	os.Exit(failRedis)
	//}

}

func checkMail(db db.DBInterface, emailAddr string) string {
	if db.Find(emailAddr) {
		return pass
	}

	return decline
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

//func PrintUsage() {
//	fmt.Println("Usage:")
//	fmt.Println("bouncer -c config.yaml file.eml")
//	fmt.Println("or")
//	fmt.Println("cat file.eml | bouncer -c config.yaml")
//	fmt.Println("or")
//	fmt.Println("use MTA pipe transport: 'command = /bin/bouncer -c /etc/bouncer.yaml'")
//	os.Exit(0)
//}
