package gormzerolog

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog"
)

type Record struct {
	occurredAt time.Time
	source     string
	duration   time.Duration
	sql        string
	values     []string
	other      []string
}

func (l *Logger) toZRFields(r *Record) {
	newLog := l.Event
	newLog.Time("occurredAt", r.occurredAt)
	newLog.Str("source", r.source)
	newLog.Dur("duration", r.duration)
	newLog.Str("sql", r.sql)
	newLog.Strs("values", r.values)
	newLog.Strs("other", r.other)
}

func (l *Logger) createLog(values []interface{}) *Record {
	ret := &Record{}
	ret.occurredAt = gorm.NowFunc()

	if len(values) > 1 {
		var level = values[0]
		ret.source = getSource(values)

		if level == "log" {
			// By default, assume this is a user log.
			// See: https://github.com/jinzhu/gorm/blob/32455088f24d6b1e9a502fb8e40fdc16139dbea8/scope.go#L96
			// If this is an error log, we set level to error.
			// See: https://github.com/jinzhu/gorm/blob/32455088f24d6b1e9a502fb8e40fdc16139dbea8/main.go#L718
			if _, ok := values[2].(error); ok {
				l.Level = zerolog.ErrorLevel
			}

			return &Record{
				other:  append(ret.other, fmt.Sprint(values[2:])),
				source: fmt.Sprintf("%v", values[1]),
			}
		} else if level == "sql" {

			l.Level = zerolog.InfoLevel

			ret.duration = getDuration(values)
			ret.values = getFormattedValues(values)
			ret.sql = getFormatSQL(values, ret.values)
			if len(values) >= 6 {
				ret.other = append(ret.other, strconv.FormatInt(values[5].(int64), 10)+" rows affected or returned ")
			}
		} else {
			ret.other = append(ret.other, fmt.Sprint(values[2:]))
		}
	}

	return ret
}
