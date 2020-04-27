package analyzer

import (
	"regexp"
	"strconv"
	"strings"
)

const (
	statusNotFound = "0.0.0"
)

//Тип баунса
//type BounceType int
//
//const (
//	Soft BounceType = 0
//	Hard BounceType = 1
//)
//
//type SMTPCode string
//
//const (
//	ServiceNotAvailable    SMTPCode = "421"
//	MailActionNotTaken     SMTPCode = "450"
//	ErrorProcessing        SMTPCode = "451"
//	InsufficientStorage    SMTPCode = "452"
//	AuthFailed             SMTPCode = "454"
//	CmdSyntaxError         SMTPCode = "500"
//	ArgumentsSyntaxError   SMTPCode = "501"
//	CmdNotImplemented      SMTPCode = "502"
//	BadCmdSequence         SMTPCode = "503"
//	CmdParamNotImplemented SMTPCode = "504"
//	BadEmailAddress        SMTPCode = "511"
//	SpamDetected           SMTPCode = "535"
//	MailboxInactive        SMTPCode = "540"
//	MessageRejected        SMTPCode = "541"
//	SpamSuspected          SMTPCode = "543"
//	MailboxUnavailable     SMTPCode = "550"
//	RecipientNotLocal      SMTPCode = "551"
//	ExceededStorageAlloc   SMTPCode = "552"
//	MailboxNameInvalid     SMTPCode = "553"
//	TransactionFailed      SMTPCode = "554"
//	TransactionProhibited  SMTPCode = "571"
//)

type Result struct {
	//Type	BounceType
	SMTPCode   int
	SMTPStatus string
	Reason     string
}

func Analyze(body []byte) Result {
	code, status, message := findBounceMessage(body)
	return Result{
		SMTPCode:   code,
		SMTPStatus: status,
		Reason:     message,
	}
}

func findBounceMessage(body []byte) (code int, status string, message string) {
	//Ценный Diagnostic-Code находится обычно в конце тела, поэтому перевернем body и приведем к нижнему регистру
	lns := strings.Split(strings.ToLower(string(body)), "\n")
	numLines := len(lns)

	var lines = make([]string, numLines)

	for i, l := range lns {
		lines[numLines-i-1] = l
	}

	//Ищем нужные вхождения
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "diagnostic-code:") {
			code, status, message = analyzeDiagCode(strings.TrimSpace(line[16:]))
		}
	}

	return code, status, message
}

func analyzeDiagCode(s string) (code int, status string, message string) {
	var (
		statusRegexp = regexp.MustCompile(`^\d\.\d\.\d+$`)
	)
	status = statusNotFound
	//вначале идет smtp;
	parts := strings.Split(strings.TrimSpace(strings.TrimPrefix(s, "smtp;")), " ")

	if len(parts) > 1 {
		if len(parts[0]) <= 3 {
			code, _ = parseCode(parts[0])

			if statusRegexp.MatchString(parts[1]) {
				status = parts[1]
				message = strings.Join(parts[2:], " ")
			} else {
				message = strings.Join(parts[1:], " ")
			}
		} else {
			allOther := parts[1:]
			dashedCode := strings.Split(parts[0], "-")

			if len(dashedCode) > 1 {
				code, _ = parseCode(dashedCode[0])

				if statusRegexp.MatchString(dashedCode[1]) {
					status = dashedCode[1]
				} else {
					allOther = append([]string{dashedCode[1]}, allOther...)
				}
			}
			message = strings.Join(allOther, " ")
		}
	} else {
		return 0, status, "Unknown diagnostic code"
	}

	return code, status, message
}

func parseCode(s string) (code int, err error) {
	if str, err := strconv.ParseInt(s, 10, 32); err == nil {
		return int(str), nil
	}

	return 0, err
}
