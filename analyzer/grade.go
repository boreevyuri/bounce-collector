package analyzer

import (
	"regexp"
	"strings"
)

//type RecordInfo struct {
//	Domain     string `json:"domain"`
//	Reason     string `json:"reason"`
//	Reporter   string `json:"reporter"`
//	SMTPCode   int    `json:"code"`
//	SMTPStatus string `json:"status"`
//	Date       string `json:"date"`
//}

const (
	MailFormatError int = 0
	DNSError        int = 0
	iCLoudFull      int = 86400 * 7
	iCloudBan       int = 0
	RateLimit       int = 0
	DomainRateLimit int = 0
	SpamBLock       int = 86400
	OverQuota       int = 86400
	Disabled        int = 86400 * 3
	NoSuchUser      int = 86400 * 7
)

var llString = []string{
	"line length exceeded",
	"line too long",
}

var lackDNSStrings = []string{
	"mx record",
	" dkim",
	" spf ",
	"find your",
}

func DetermineReason(r *RecordInfo) (ttl int, err error) {
	reason := normalizeMessage(r.Reason)

	switch {
	case icloudOverquota(reason, r):
		ttl = iCLoudFull
	case findSubstring(reason, llString):
		ttl = MailFormatError
	case findSubstring(reason, lackDNSStrings):
		ttl = DNSError
	default:
		ttl = 0
	}

	return 0, nil
}

//Нужен рефакторинг. Здесь адова чертовщина.
func normalizeMessage(reason string) string {
	bogusSymbols := [4]string{"'", "\"", ",", ":"}
	emailRegexp := regexp.MustCompile(`<?\b[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}\b>?`)
	bogusSMTPReason := regexp.MustCompile(`#(\d\.\d\.\d+)`)
	space := regexp.MustCompile(`\s+`)

	//Не забываем убирать пробелы
	reason = strings.TrimSpace(strings.ToLower(reason))
	//Убираем дописанное нашим MTA
	reason = strings.TrimSuffix(reason, "retry timeout exceeded")

	//Вырезаем точки, кавычки, запятые и двоеточия
	for _, s := range bogusSymbols {
		reason = strings.ReplaceAll(reason, s, "")
	}
	//Вырезаем email адреса
	reason = emailRegexp.ReplaceAllString(reason, "")
	//Меняем дефисы на пробелы
	reason = strings.ReplaceAll(reason, "-", " ")
	//Схлопываем лишние пробелы
	reason = space.ReplaceAllString(reason, " ")
	//Приводим в порядок коды вида #5.1.1
	reason = bogusSMTPReason.ReplaceAllString(reason, "${1}")
	//Еще раз тримим на пробелы
	reason = strings.TrimSpace(reason)

	return reason
}

//Определение iCloud overquota.
func icloudOverquota(s string, r *RecordInfo) bool {
	return (r.Domain == "icloud.com" || r.Domain == "me.com") &&
		r.SMTPCode == 450 &&
		r.SMTPStatus == "4.2.2" &&
		strings.Contains(s, "overquota")
}

func findSubstring(s string, arr []string) bool {
	for _, i := range arr {
		if strings.Contains(s, i) {
			return true
		}
	}
	return false
}
