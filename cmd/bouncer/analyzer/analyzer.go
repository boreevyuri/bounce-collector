package analyzer

import (
	"errors"
	"mime"
	"mime/multipart"
	"net/mail"
	"regexp"
	"strconv"
	"strings"
)

const (
	contentTypeHeader   = "Content-Type"
	expectedContentType = "multipart/report"
	expectedReportType  = "delivery-status"
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

// Result struct describes result.
type Result struct {
	//Type	BounceType
	SMTPCode   int
	SMTPStatus string
	Reason     string
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

// Analyze - analyses body for error messages in it.
func Analyze(body []byte) Result {
	if res, err := findBounceMessage(body); err == nil {
		return res
	}

	return Result{
		SMTPCode:   0,
		SMTPStatus: "0.0.0",
		Reason:     "Unable to find reason",
	}
}

func NewAnalyze(m *mail.Message) (d string, err error) {

	boundary, err := findBoundary(&m.Header)
	bodyReader := multipart.NewReader(m.Body, boundary)

	return
}

func findBoundary(h *mail.Header) (boundary string, err error) {
	mediaType, params, err := mime.ParseMediaType(h.Get(contentTypeHeader))
	if err != nil {
		return boundary, err
	}
	if mediaType == expectedContentType && params["report-type"] == expectedReportType {
		if boundary, ok := params["boundary"]; ok {
			return boundary, nil
		}
	}

	return boundary, errors.New("this is not exim report")
}

func findBounceMessage(body []byte) (res Result, err error) {
	//Ценный Diagnostic-Code находится обычно в конце тела, поэтому перевернем body и приведем к нижнему регистру
	lns := strings.Split(strings.ToLower(string(body)), "\n")
	numLines := len(lns)
	lines := make([]string, numLines)

	for i, l := range lns {
		lines[numLines-i-1] = l
	}

	//Ищем нужные вхождения
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

func analyzeDiagCode(s string) (res Result, err error) {
	var (
		statusRegexp = regexp.MustCompile(`^\d\.\d\.\d+$`)
	)

	//status := statusNotFound

	//вначале идет smtp;
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
