package analyzer

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	statusNotFound    = "0.0.0"
	BadEmailAddress   = "5.1.1"
	diagCodeHeader    = "diagnostic-code:"
	SMTPCodeLen       = 3
	unrouteableString = "unrouteable address"
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
	if res, err := findBounceMessage(body); err == nil {
		return res
	}

	return Result{
		SMTPCode:   0,
		SMTPStatus: "0.0.0",
		Reason:     "Unable to find reason",
	}
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
			res.SMTPStatus = BadEmailAddress
			res.Reason = "Unrouteable address"

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
		if len(parts[0]) <= SMTPCodeLen {
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
