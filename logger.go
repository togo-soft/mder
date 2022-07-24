package main

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func init() {
	logger.SetReportCaller(true)
	logger.SetFormatter(&EasyFormatter{})
}

type EasyFormatter struct{}

func (receiver *EasyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var output bytes.Buffer
	output.WriteString(entry.Level.String()[:4])
	output.WriteString("|")
	var filenames = strings.Split(entry.Caller.File, "/")
	filename := filenames[len(filenames)-1]
	output.WriteString(fmt.Sprintf("%s:%d", filename, entry.Caller.Line))
	output.WriteString("|")
	output.WriteString(entry.Caller.Function)
	output.WriteString("|")
	for k, val := range entry.Data {
		output.WriteString(fmt.Sprintf("%s=%v|", k, val))
	}
	output.WriteString(" " + entry.Message)
	output.WriteRune('\n')
	return output.Bytes(), nil
}
