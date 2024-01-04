// Package alog contains some methods from the glog package, but also logs errors to the CloudMessage
package alog

import (
	"birnenlabs.com/go/lib/automate"
	"fmt"
	"github.com/golang/glog"
)

var cloudMsg *automate.CloudMessage

func init() {
	c, err := automate.Create()
	if err == nil {
		cloudMsg = c
	} else {
		cloudMsg = nil
		glog.Errorf("Could not create CloudMessage instance: %v.", err)
	}
}

// When appName is set, the formatted payload will be used for the cloud messages:
// $from [$hostname]|$priority|$message
func UseFormattedPayload(appName string) {
	cloudMsg.UseFormattedPayload(appName)
}

// Some methods that are defined in glog.
// Error*, Fatal* and Exit* will also send cloud message.

func Info(args ...interface{}) {
	glog.InfoDepth(1, args...)
}

func Infof(format string, args ...interface{}) {
	glog.InfoDepth(1, fmt.Sprintf(format, args...))
}

func Warning(args ...interface{}) {
	glog.WarningDepth(1, args...)
}

func Warningf(format string, args ...interface{}) {
	glog.WarningDepth(1, fmt.Sprintf(format, args...))
}

func Error(args ...interface{}) {
	glog.ErrorDepth(1, sendUrgent(args...))
}

func Errorf(format string, args ...interface{}) {
	glog.ErrorDepth(1, sendUrgentf(format, args...))
}

func Fatal(args ...interface{}) {
	glog.FatalDepth(1, sendUrgent(args...))
}

func Fatalf(format string, args ...interface{}) {
	glog.FatalDepth(1, sendUrgentf(format, args...))
}

func Exit(args ...interface{}) {
	glog.ExitDepth(1, sendUrgent(args...))
}

func Exitf(format string, args ...interface{}) {
	glog.ExitDepth(1, sendUrgentf(format, args...))
}

func Flush() {
	glog.Flush()
}

// Also support V(x) as in glog:

type Verbose bool

func V(level int) Verbose {
	return Verbose(glog.V(glog.Level(level)))
}

func (v Verbose) Info(args ...interface{}) {
	if v {
		glog.InfoDepth(1, args...)
	}
}

func (v Verbose) Infof(format string, args ...interface{}) {
	if v {
		glog.InfoDepth(1, fmt.Sprintf(format, args...))
	}
}

// Additional methods that logs to the info level and sends default priority cloud message

func InfoSend(args ...interface{}) {
	glog.InfoDepth(1, send(args...))
}

func InfoSendf(format string, args ...interface{}) {
	glog.InfoDepth(1, sendf(format, args...))
}

func InfoSendUrgent(args ...interface{}) {
	glog.InfoDepth(1, sendUrgent(args...))
}

func InfoSendUrgentf(format string, args ...interface{}) {
	glog.InfoDepth(1, sendUrgentf(format, args...))
}

// Internal send* methods.

func send(args ...interface{}) string {
	msg := fmt.Sprint(args...)
	cloudMsg.Send(msg)
	return msg
}

func sendf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	cloudMsg.Send(msg)
	return msg
}

func sendUrgent(args ...interface{}) string {
	msg := fmt.Sprint(args...)
	cloudMsg.SendUrgent(msg)
	return msg
}

func sendUrgentf(format string, args ...interface{}) string {
	msg := fmt.Sprintf(format, args...)
	cloudMsg.SendUrgent(msg)
	return msg
}
