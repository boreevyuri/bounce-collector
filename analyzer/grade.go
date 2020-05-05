package analyzer

import (
	"regexp"
	"strings"
	"time"
)

const (
	MailFormatError = 0 * time.Second
	DNSError        = 0 * time.Second
	iCLoudFull      = 7 * 24 * time.Hour
	iCloudBan       = 0 * time.Second
	RateLimit       = 0 * time.Second
	SpamBLock       = 0 * time.Second
	OverQuota       = 24 * time.Hour
	Disabled        = 15 * 24 * time.Hour
	NoSuchUser      = 30 * 24 * time.Hour
	NoSuchDomain    = 90 * 24 * time.Hour
	//DomainRateLimit int = 0
)

func DetermineReason(r *RecordInfo) (ttl time.Duration, err error) {
	switch r.Reason {
	case unrouteableString:
		ttl = NoSuchDomain
	default:
		ttl, err = lastHopeDetermine(r)
	}

	return ttl, err
}

func lastHopeDetermine(r *RecordInfo) (ttl time.Duration, err error) {
	var (
		llString          = []string{"line length exceeded", "line too long"}
		lackDNSStrings    = []string{"mx record", " dkim", " spf ", "find your"}
		proofpointStrings = []string{`^.+ipcheck\.proofpoint\.com.+$`}
		rateLimitMessage  = []string{`^.*too many.*$`, `^.*rate.*$`}
		spamBlockMessage  = []string{
			`^.+(spam|dnsbl|abus|reputat|policy|blacklis|securit|tenantattribution|banned|complain|outside|prohibit|rdeny|allow|aiuthentic|permiso).+$`, //nolint:lll
			`^.+sender.+denied.*$`, `^.*relay access denied.*$`, `^.*service refuse.*$`, `^.*rejected by recipient.*$`,
			`^not$`,
		}
		overQuotaMessage = []string{`^.*quota.*$`, `^.*mailbox.+limit.*$`}
		disabledMessage  = []string{`^.*(inactive|blocked|expired|suspend|frozen|disabled|locked|enable).*$`}
		absentMessage    = []string{`^.*(invalid|unknown|rejected|bad|unavailable).*$`, `^.*(no such).*$`, `^.*not.*$`,
			`^.*no mailbox.*$`, `^.*no longer available.*$`, `^.*unrouteable address.*$`,
			`^.*delivery error dd.*$`, `^.*server disconnected.*$`,
		}
	)

	if failedSpamDelivery(r) {
		ttl = SpamBLock
	} else {
		reason := normalizeMessage(r.Reason)

		switch {
		case icloudOverquota(reason, r):
			ttl = iCLoudFull
		case findSubstring(reason, llString):
			ttl = MailFormatError
		case findSubstring(reason, lackDNSStrings):
			ttl = DNSError
		case stringMatchAnyRegex(reason, proofpointStrings):
			ttl = iCloudBan
		case stringMatchAnyRegex(reason, rateLimitMessage):
			ttl = RateLimit
		case stringMatchAnyRegex(reason, spamBlockMessage):
			ttl = SpamBLock
		case stringMatchAnyRegex(reason, overQuotaMessage):
			ttl = OverQuota
		case stringMatchAnyRegex(reason, disabledMessage):
			ttl = Disabled
		case stringMatchAnyRegex(reason, absentMessage):
			ttl = NoSuchUser
		default:
			ttl = 0
		}
	}

	return ttl, nil
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

func stringMatchAnyRegex(ss string, regexes []string) bool {
	for _, re := range regexes {
		result, err := regexp.MatchString(re, ss)
		if err != nil {
			panic(err)
		}

		if result {
			return true
		}
	}

	return false
}

func findSubstring(s string, arr []string) bool {
	for _, i := range arr {
		if strings.Contains(s, i) {
			return true
		}
	}

	return false
}

// yandex.ru 451 4.7.1 - spam block.
func failedSpamDelivery(r *RecordInfo) bool {
	return r.SMTPCode == 451 && r.SMTPStatus == "4.7.1"
}
