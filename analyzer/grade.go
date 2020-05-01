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

func DetermineReason(r *RecordInfo) (ttl int, err error) {
	reason := normalizeMessage(r.Reason)

	return 0, nil
}

//Нужен рефакторинг. Здесь адов пипец
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
