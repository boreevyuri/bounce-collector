package analyzer

import (
	"bounce-collector/cmd/bouncer/database"
	"errors"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

const (
	statusNotFound      = "0.0.0"
	badEmailAddressCode = "5.1.1"
	diagCodeHeader      = "diagnostic-code:"
	smtpCodeLength      = 3
	unrouteableString   = "unrouteable address"
	nonExistentString   = "all relevant mx records point to non-existent hosts"
	nonExistSMTPString  = "an mx or srv record indicated no smtp service"
)

// bodyData struct describes result.
type bodyData struct {
	// Type	BounceType
	Reason     string
	SMTPCode   int
	SMTPStatus string
}

// RecordInfo struct describes record for every email putted in redis.
type RecordInfo struct {
	Domain     string `json:"domain"`
	Reason     string `json:"reason"`
	Reporter   string `json:"reporter"`
	SMTPCode   int    `json:"code"`
	SMTPStatus string `json:"status"`
	Date       string `json:"date"`
}

type MailData struct {
	mail *mail.Message
	res  chan database.RecordPayload
}

// MailAnalyzer is a channel for analyzer.
type MailAnalyzer chan MailData

// New - creates new analyzer instance.
func New() *MailAnalyzer {
	mailAnalyzer := make(MailAnalyzer)
	go mailAnalyzer.Analyze()
	return &mailAnalyzer
}

// Analyze - analyzes body for error messages in it.
func (ma *MailAnalyzer) Analyze() {
	for cmd := range *ma {
		cmd.res <- ma.doAnalyze(cmd.mail)
	}
}

// doAnalyze - analyzes body for error messages in it.
func (ma *MailAnalyzer) doAnalyze(m *mail.Message) database.RecordPayload {
	// Get data from mail headers
	rcpt := strings.ToLower(m.Header.Get("X-Failed-Recipients"))
	date := m.Header.Get("Date")
	from := parseFrom(m.Header.Get("From"))

	// Get data from mail body
	//body, _ := io.ReadAll(m.Body)
	//smtpCode, smtpStatus, reason := ma.Parse(body)

	messageInfo := RecordInfo{
		Domain:     getDomainFromAddress(rcpt),
		SMTPCode:   0,
		SMTPStatus: "0.0.0",
		Reason:     "Unable to find reason",
		Date:       date,
		Reporter:   from,
	}

	return database.RecordPayload{
		Key:   rcpt,
		Value: messageInfo,
		TTL:   0,
	}
}

// Parse - analyzes body for error messages in it.
func (ma *MailAnalyzer) Parse(body []byte) (int, string, string) {

}

// parseDiagCode - parses diagnostic code from body.

// Rotate and decapitalize body for more comfortable string searching.
func (ma *MailAnalyzer) rotateAndDecapitalize(body []byte) []string {
	lns := strings.Split(strings.ToLower(string(body)), "\n")
	numLines := len(lns)
	lines := make([]string, numLines)

	for i, l := range lns {
		lines[numLines-i-1] = l
	}

	return lines
}

// ===================
// Analyze - analyzes body for error messages in it.
func Analyze(body []byte) (int, string, string) {
	if res, err := findBounceMessage(body); err == nil {
		return res.SMTPCode, res.SMTPStatus, res.Reason
	}

	return 0, "0.0.0", "Unable to find reason"
}

func findBounceMessage(body []byte) (res bodyData, err error) {
	// Ценный Diagnostic-Code находится обычно в конце тела, поэтому перевернем body и приведем к нижнему регистру
	lns := strings.Split(strings.ToLower(string(body)), "\n")
	numLines := len(lns)
	lines := make([]string, numLines)

	for i, l := range lns {
		lines[numLines-i-1] = l
	}

	// Ищем нужные вхождения
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Приоритет на Diagnostic-Code
		if strings.HasPrefix(line, diagCodeHeader) {
			res, err = analyzeDiagCode(strings.TrimSpace(line[len(diagCodeHeader):]))
			if err != nil {
				break
			}
		}

		// Проверка на наличие Unrouteable address
		if strings.EqualFold(line, unrouteableString) {
			res.SMTPCode = 550
			res.SMTPStatus = badEmailAddressCode
			res.Reason = unrouteableString

			break
		}

		// Проверка на MX records point to non-existent hosts
		if strings.EqualFold(line, nonExistentString) {
			res.SMTPCode = 550
			res.SMTPStatus = badEmailAddressCode
			res.Reason = unrouteableString

			break
		}

		// Проверка на an MX or SRV record indicated no SMTP service
		if strings.EqualFold(line, nonExistSMTPString) {
			res.SMTPCode = 550
			res.SMTPStatus = badEmailAddressCode
			res.Reason = unrouteableString

			break
		}
	}

	return res, err
}

func analyzeDiagCode(s string) (res bodyData, err error) {
	var (
		statusRegexp = regexp.MustCompile(`^\d\.\d\.\d+$`)
	)

	// status := statusNotFound

	// вначале идет smtp;
	parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(s, "smtp;")), " ")

	if len(parts) > 1 {
		if len(parts[0]) <= smtpCodeLength {
			res.SMTPCode = parseCode(parts[0])

			if statusRegexp.MatchString(parts[1]) {
				res.SMTPStatus = parts[1]
				res.Reason = strings.Join(parts[2:], " ")
			} else {
				res.Reason = strings.Join(parts[1:], " ")
			}
		} else {
			allOther := parts[1:]
			dashedCode := strings.Split(parts[0], "-")

			if len(dashedCode) > 1 {
				res.SMTPCode = parseCode(dashedCode[0])

				if statusRegexp.MatchString(dashedCode[1]) {
					res.SMTPStatus = dashedCode[1]
				} else {
					res.SMTPStatus = statusNotFound
					allOther = append([]string{dashedCode[1]}, allOther...)
				}
			}
			res.Reason = strings.Join(allOther, " ")
		}
	} else {
		err = errors.New("unable to parse diagnostic code")
	}

	return res, err
}

func parseCode(s string) (code int) {
	if str, err := strconv.ParseInt(s, 10, 32); err == nil {
		return int(str)
	}

	return 0
}

// parseFrom - parses from address from string.
func parseFrom(s string) string {
	e, err := mail.ParseAddress(s)
	if err != nil {
		return "unknown@unknown.tld"
	}

	return e.Address
}

// getDomainFromAddress - returns domain from mail address.
func getDomainFromAddress(addr string) string {
	a := strings.Split(addr, "@")
	if len(a) > 1 {
		return a[1]
	}

	return "unknown.tld"
}
