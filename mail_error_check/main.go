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
const oneMonth = 30 * 24 * 3600

type Config struct {
	LastRun int64
}

var dryRun = flag.Bool("dryrun", false, "Dry run")

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "true")
	defer glog.Flush()

	glog.Infof("Dry run mode: %v", *dryRun)

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

	// Load configuration
	var config Config
	err = conf.LoadConfigFromFile(appName, &config)
	if err != nil {
		// Not exiting here, let's just read all messages
		glog.Warningf("Last run time not found (%s). Listing all messages.", err)
	}

	now := time.Now().Unix()

	// Process our events first...
	err = processEvents(m, max(config.LastRun, now-oneMonth), now)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exit("Could not process our emails", err)
	}

	// ...and save our events state
	config.LastRun = now
	if *dryRun {
		glog.Infof("DRY RUN: would save config: %+v", config)
	} else {
		err = conf.SaveConfigToFile(appName, &config)
		if err != nil {
			cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
			glog.Errorf("Could not save last run time to file: %s", err)
		}
	}

	cloudMessage.SendFormattedCloudMessageToDefault(appName, "Done", 0)
}

func processEvents(m *mailgun.Mailgun, begin, end int64) error {
	glog.Infof("Processing emails between %d and %d", begin, end)
	items, err := m.ListAllEvents(begin, end)
	if err != nil {
		return err
	}

	for _, item := range items {
		// Ignoring second and following attempts of temporary failure
		if item.Severity == temporarySeverity && item.DeliveryStatus.AttemptNo > 1 {
			glog.Infof("Ignoring %d attempt of %s error from: %s to: %s", item.DeliveryStatus.AttemptNo, temporarySeverity, item.From(), item.To())
			continue
		}

		email := mailgun.Email{
			From:      m.MailerDaemon(),
			Text:      generateBounceEmailText(item),
			Subject:   "Re: " + item.Message.Headers.Subject,
			Reference: "<" + item.Message.Headers.MessageId + ">",
		}

		if m.IsInMyDomain(item.From()) {
			// Always inform about actions originating in own domain
			email.To = item.From()
		} else if item.Event == mailgun.EventFailed {
			// Send email to postmaster for other domain's failures
			email.To = m.CreateAddress("postmaster")
		} else {
			// Ignore everything else
			glog.Infof("Ignoring stored/rejected event not originating from domain. From %s, To: %s", item.From(), item.To())
			continue
		}

		if *dryRun {
			glog.Infof("DRY RUN: would send email:\n %+v", email)
		} else {
			err = m.SendBounceEmail(email, item.To() /*failedRecipient*/)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generateBounceEmailText(item mailgun.Item) string {
	isTemp := (item.Severity == temporarySeverity)

	result := fmt.Sprintf("Mail Delivery %s Failure.\n\nThis is an automatically generated Delivery Status Notification.\n\n", strings.Title(item.Severity))

	if isTemp {
		result = result + "THIS IS A WARNING MESSAGE ONLY.\nYOU DO NOT NEED TO RESEND YOUR MESSAGE.\nSERVER WILL RETRY FOR THE NEXT 12 HOURS.\n\n"
	}

	result = result + fmt.Sprintf("Delivery of the following message:\n\n\tFrom: %s\n\tTo: %s\n\n", item.From(), item.To())

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
