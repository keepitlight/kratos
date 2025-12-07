package log

import (
	"encoding/json"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	alarmKey   = "Alarm"
	alarmTitle = "告警提醒"
)

// SeverityLevel 采集日志告警级别
type SeverityLevel string

const (
	SeverityLow    SeverityLevel = "low"
	SeverityMiddle SeverityLevel = "middle"
	SeverityHigh   SeverityLevel = "high"
)

type Data struct {
	Severity SeverityLevel `json:"Severity,omitempty"`
	Title    string        `json:"Title,omitempty"`
	Details  string        `json:"Details,omitempty"`
}

func alarm(log *log.Helper, s SeverityLevel, msg string) {
	log.Errorw(alarmKey, Data{
		Severity: s,
		Title:    alarmTitle,
		Details:  msg,
	})
}

func combo(msg string, err error) string {
	if err != nil {
		return msg + err.Error()
	}
	return msg
}

// Report 告警
func Report(log *log.Helper, msg string, err error) {
	alarm(log, SeverityLow, combo(msg, err))
}

// Urgent 严重告警
func Urgent(log *log.Helper, err error) {
	alarm(log, SeverityMiddle, err.Error())
}

// Fatal 致命告警
func Fatal(log *log.Helper, msg string, err error) {
	alarm(log, SeverityHigh, combo(msg, err))
}

// ReportF 告警
func ReportF(log *log.Helper, format string, args ...any) {
	alarm(log, SeverityLow, fmt.Sprintf(format, args...))
}

// UrgentF 严重告警
func UrgentF(log *log.Helper, format string, args ...any) {
	alarm(log, SeverityMiddle, fmt.Sprintf(format, args...))
}

// UrgentW 严重告警
func UrgentW(log *log.Helper, args any) {
	v, _ := json.Marshal(args)
	log.Errorw(alarmKey, Data{
		Severity: SeverityMiddle,
		Title:    alarmTitle,
		Details:  string(v),
	})
}

// FatalF 致命告警
func FatalF(log *log.Helper, format string, args ...any) {
	alarm(log, SeverityHigh, fmt.Sprintf(format, args...))
}
