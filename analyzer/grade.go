package analyzer

import (
	"regexp"
	"strings"
)

const (
	MailFormatError int = 0
	DNSError        int = 0
	iCLoudFull      int = 86400 * 7
	iCloudBan       int = 0
	RateLimit       int = 0
	SpamBLock       int = 0
	OverQuota       int = 86400
	Disabled        int = 86400 * 15
	NoSuchUser      int = 86400 * 30
	NoSuchDomain    int = 86400 * 90
	//DomainRateLimit int = 0
)

func DetermineReason(r *RecordInfo) (ttl int, err error) {
	switch r.Reason {
	case unrouteableString:
		ttl = NoSuchDomain
	default:
		ttl, err = lastHopeDetermine(r)
	}

	return ttl, err
}

func lastHopeDetermine(r *RecordInfo) (ttl int, err error) {
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
