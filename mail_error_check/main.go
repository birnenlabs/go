package main

import (
	"birnenlabs.com/automate"
	"birnenlabs.com/conf"
	"birnenlabs.com/mailgun"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"strings"
	"time"
)

const appName = "mail_error_check"

type Config struct {
	LastRun int64
}

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "true")
	defer glog.Flush()

	// Create notifier first
	cloudMessage, err := automate.Create()
	if err != nil {
		// Not exiting here, continue without cloud notifier.
		glog.Errorf("Could not create cloud message: %v", err)
	}

	glog.Infof("Creating new mailgun instance")
	m, err := mailgun.New()
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exit("Could not create mailgun instance:", err)
	}

	var config Config
	err = conf.LoadConfigFromFile(appName, &config)
	if err != nil {
		// Not exiting here, let's just read all messages
		glog.Warningf("Last run time not found (%s). Listing all messages.", err)
	}

	now := time.Now().Unix()
	glog.Infof("Listing emails between %d and %d", now, config.LastRun)

	items, err := m.ListFailedEventsTimeRange(now, config.LastRun)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exit("Could not list failed events:", err)
	}

	for _, item := range items {
		glog.Infof("Message from: %v to %v - code: %d", item.Envelope.Sender, item.Envelope.Targets, item.DeliveryStatus.Code)

		email := mailgun.Email{
			From:      item.Envelope.Targets,
			To:        []string{item.Envelope.Sender},
			Subject:   "Re: " + item.Message.Headers.Subject,
			Text:      generateErrorEmailText(item),
			Reference: "<" + item.Message.Headers.MessageId + ">",
		}

		err = m.SendBounceEmail(email)
		if err != nil {
			cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
			glog.Exit("Could not send email:", err)
		}
	}

	config = Config{
		LastRun: now,
	}
	err = conf.SaveConfigToFile(appName, &config)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Errorf("Could not save last run time to file: %s", err)
	}

	cloudMessage.SendFormattedCloudMessageToDefault(appName, fmt.Sprintf("Processed %d messages", len(items)), 0)
}

func generateErrorEmailText(item mailgun.Item) string {
	var tmpl string
	if item.Severity == "permanent" {
		tmpl = permanentErrorTmpl
	} else {
		tmpl = temporaryErrorTmpl
	}

	reason := ""
	if len(item.DeliveryStatus.Message) > 0 {
		reason = item.DeliveryStatus.Message
	}
	if len(item.DeliveryStatus.Description) > 0 && item.DeliveryStatus.Message != item.DeliveryStatus.Description {
		reason = reason + "\n" + item.DeliveryStatus.Description
	}

	return fmt.Sprintf(tmpl, strings.Title(item.Severity), item.Envelope.Targets, item.DeliveryStatus.AttemptNo, item.DeliveryStatus.Code, item.Reason, reason)
}
