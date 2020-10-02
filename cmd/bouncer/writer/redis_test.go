package writer

import (
	"bounce-collector/cmd/bouncer/analyzer"
	"bounce-collector/cmd/bouncer/config"
	"github.com/alicebob/miniredis"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

var (
	client     ProcessRedis
	testRecord = Record{
		Rcpt: "aaa@bbb.ru",
		TTL:  time.Duration(60) * time.Second,
		Info: analyzer.RecordInfo{
			Domain:     "bbb.ru",
			Reason:     "failed",
			Reporter:   "server",
			SMTPCode:   550,
			SMTPStatus: "550",
			Date:       "01-02-2003",
		},
	}
)

//var (
//	key = "key"
//	val = "val"
//)

func TestMain(m *testing.M) {
	mr, err := miniredis.Run()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database", err)
	}

	client = New(config.RedisConfig{
		Addr: mr.Addr(),
	})

	code := m.Run()
	os.Exit(code)
}

func TestNew(t *testing.T) {
	assert.Implements(t, (*ProcessRedis)(nil), client)
}

func TestProcessRedis_Insert(t *testing.T) {
	//exp := time.Duration(0)

	res := client.Insert(testRecord)
	assert.Equal(t, true, res)
}

func TestProcessRedis_Find(t *testing.T) {
	res := client.Find(testRecord.Rcpt)
	assert.Equal(t, true, res)
}
