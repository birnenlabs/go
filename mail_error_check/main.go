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
const temporarySeverity = "temporary"

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
	from := max(config.LastRun, now - 2592000) // 2592000 == 30*24*3600

	glog.Infof("Listing emails between %d and %d", now, from)

	items, err := m.ListFailedEventsTimeRange(now, from)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exit("Could not list failed events:", err)
	}

	for _, item := range items {
		if item.Severity == temporarySeverity && item.DeliveryStatus.AttemptNo > 1 {
			glog.Infof("Ignoring %d attempt of %s error from: %s to: %s", item.DeliveryStatus.AttemptNo, temporarySeverity, item.Envelope.Sender, item.Envelope.Targets)
			continue
		}
		glog.Infof("Will send warning for message from: %s to: %s.", item.Envelope.Sender, item.Envelope.Targets)

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
	isTemp := (item.Severity == temporarySeverity)

	result := fmt.Sprintf("Mail Delivery %s Failure.\n\nThis is an automatically generated Delivery Status Notification.\n\n", strings.Title(item.Severity))

	if isTemp {
		result = result + "THIS IS A WARNING MESSAGE ONLY.\nYOU DO NOT NEED TO RESEND YOUR MESSAGE.\n\n"
	}

	result = result + fmt.Sprintf("Delivery to the following recipients:\n\n\t\t%s\n\n", item.Envelope.Targets)

	if isTemp {
		result = result + "has been delayed.\n\n\nTechnical details:\n"
	} else {
		result = result + fmt.Sprintf("failed permanently after %d attempts.\n\n\nTechnical details:\n", item.DeliveryStatus.AttemptNo)
	}

	result = result + fmt.Sprintf("Timestamp: %s\nSeverity: %s\nSMTP code: %d\nReason: %s\n\n", time.Unix(int64(item.Timestamp), 0).UTC(), item.Severity, item.DeliveryStatus.Code, item.Reason)
	if len(item.DeliveryStatus.Message) > 0 {
		result = result + item.DeliveryStatus.Message + "\n\n"
	}
	if len(item.DeliveryStatus.Description) > 0 && item.DeliveryStatus.Message != item.DeliveryStatus.Description {
		result = result + item.DeliveryStatus.Description + "\n\n"
	}

	return result
}

func max(x, y int64) int64 {
    if x > y {
        return x
    }
    return y
}
