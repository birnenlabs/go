package main

import (
	"birnenlabs.com/automate"
	"birnenlabs.com/conf"
	"birnenlabs.com/mailgun"
	"flag"
	"fmt"
	"github.com/golang/glog"
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

	itemsMap := m.GroupItems(items)
	for headers, messages := range itemsMap {
		if len(messages) == 0 {
			errTxt := "Zero groupped messages for message ID. This is most likely application bug."
			cloudMessage.SendFormattedCloudMessageToDefault(appName, errTxt, 1)
			glog.Exit(errTxt)
		}
		// Envelope should be the same for all messages, and we just checked if len(messages) is not zero.
		envelope := messages[0].Envelope
		glog.Infof("%d messages from: %v to %v", len(messages), headers.From, headers.To)

		email := mailgun.Email{
			From:      envelope.Targets,
			To:        []string{envelope.Sender},
			Subject:   "Re: " + headers.Subject,
			Text:      generateErrorEmailText(messages),
			Reference: "<" + headers.MessageId + ">",
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

func generateErrorEmailText(items []mailgun.Item) string {
	if len(items) == 0 {
		return ""
	}

	tmpl := temporaryErrorTmpl
	for _, item := range items {
		if item.Severity == "permanent" {
			tmpl = permanentErrorTmpl
			break
		}
	}

	details := ""
	for _, item := range items {
		reason := ""
		if len(item.DeliveryStatus.Message) > 0 {
			reason = item.DeliveryStatus.Message
		}
		if len(item.DeliveryStatus.Description) > 0 && item.DeliveryStatus.Message != item.DeliveryStatus.Description {
			reason = reason + "\n" + item.DeliveryStatus.Description
		}
		details = details + fmt.Sprintf(technicalDetailsTmpl, item.DeliveryStatus.AttemptNo, time.Unix(int64(item.Timestamp), 0).UTC(), item.Severity, item.DeliveryStatus.Code, item.Reason, reason)
	}

	return fmt.Sprintf(tmpl, items[0].Message.Headers.To, details)
}
