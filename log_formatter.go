package main

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/mgutz/ansi"
)

type colorFunc func(string) string

var errCol colorFunc
var warnCol colorFunc
var infoCol colorFunc
var debugCol colorFunc
var bold colorFunc

func init() {
	errCol = ansi.ColorFunc("white+b:red")
	warnCol = ansi.ColorFunc("208+b")
	infoCol = ansi.ColorFunc("white+b")
	debugCol = ansi.ColorFunc("white")
	bold = ansi.ColorFunc("white+b")
}

type TextFormatter struct{}

func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var keys []string = make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	b := &bytes.Buffer{}

	prefixFieldClashes(entry.Data)

	f.printColored(b, entry, keys)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *TextFormatter) printColored(b *bytes.Buffer, entry *logrus.Entry, keys []string) {
	var cf colorFunc
	switch entry.Level {
	case logrus.DebugLevel:
		cf = debugCol
	case logrus.InfoLevel:
		cf = infoCol
	case logrus.WarnLevel:
		cf = warnCol
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		cf = errCol
	default:
		cf = func(s string) string { return s }
	}

	levelText := "[" + strings.ToUpper(entry.Level.String())[0:4] + "]"

	fmt.Fprintf(b, "%s %-44s ", cf(levelText), entry.Message)
	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " %s=%+v", bold(k), v) // TODO: clamp val size
	}
}

func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return false
		}
	}
	return true
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {

	b.WriteString(key)
	b.WriteByte('=')

	switch value := value.(type) {
	case string:
		if needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	case error:
		errmsg := value.Error()
		if needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	default:
		fmt.Fprint(b, value)
	}

	b.WriteByte(' ')
}

func prefixFieldClashes(data logrus.Fields) {
	_, ok := data["time"]
	if ok {
		data["fields.time"] = data["time"]
	}

	_, ok = data["msg"]
	if ok {
		data["fields.msg"] = data["msg"]
	}

	_, ok = data["level"]
	if ok {
		data["fields.level"] = data["level"]
	}
}
