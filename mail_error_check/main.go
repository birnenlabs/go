package main

import (
	glog "birnenlabs.com/lib/alog"
	"birnenlabs.com/lib/conf"
	"birnenlabs.com/lib/mailgun"
	"flag"
	"fmt"
	"strings"
	"time"
)

const appName = "mail_error_check"
const temporarySeverity = "temporary"
const oneMonth = 30 * 24 * 3600

type State struct {
	LastRun int64
}

var dryRun = flag.Bool("dryrun", false, "Dry run")

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "true")
	defer glog.Flush()

	glog.UseFormattedPayload(appName)

	glog.Infof("Dry run mode: %v", *dryRun)

	glog.Infof("Creating new mailgun instance")
	m, err := mailgun.New()
	if err != nil {
		glog.Exit("Could not create mailgun instance:", err)
	}

	// Load configuration
	var state State
	err = conf.LoadConfigFromFile(appName, &state)
	if err != nil {
		// Not exiting here, let's just read all messages
		glog.Warningf("Last run time not found (%s). Listing all messages.", err)
	}

	// Load rules
	var config Config
	err = conf.LoadConfigFromJson(appName, &config)
	if err != nil {
		// Not exiting here, let's just read all messages
		glog.Warningf("Could not load config: %v.", err)
	}

	now := time.Now().Unix()

	err = processEvents(m, config.Rules, max(state.LastRun, now-oneMonth), now)

	if err != nil {
		glog.Exit("Could not process our emails", err)
	}

	state.LastRun = now
	if *dryRun {
		glog.Infof("DRY RUN: would save state: %+v", state)
	} else {
		err = conf.SaveConfigToFile(appName, &state)
		if err != nil {
			glog.Errorf("Could not save last run time to file: %s", err)
		}
	}

	// now/60 == minutes from epoch
	// 24 *60 == minutes in day
	if (now/60)%(24*60) < 10 {
		// Send cloud message only between 0:00 and 0:09
		glog.InfoSend("Done")
	}
}

func processEvents(m *mailgun.Mailgun, rules []Rule, begin, end int64) error {
	glog.Infof("Processing emails between %d and %d", begin, end)
	items, err := m.ListAllEvents(begin, end)
	if err != nil {
		return err
	}

	for _, item := range items {
		// Ignoring second and following attempts of temporary failure
		if item.Severity == temporarySeverity && item.DeliveryStatus.AttemptNo > 1 {
			glog.Infof("Ignore %s->%s: %d attempt, severity: %s", item.From(), item.To(), item.DeliveryStatus.AttemptNo, temporarySeverity)
			continue
		}

		matched := false
		for _, rule := range rules {
			if matches(item, rule.Match) {
				glog.Infof("Match %s->%s: %v -> %v", item.From(), item.To(), rule.Match, rule.Action)
				matched = true

				if rule.Action.NotifyPostmaster {

					email := mailgun.Email{
						From:      m.MailerDaemon(),
						To:        m.CreateAddress("postmaster"),
						Text:      generateBounceEmailText(item),
						Subject:   "Re: " + item.Message.Headers.Subject,
						Reference: "<" + item.Message.Headers.MessageId + ">",
					}

					if *dryRun {
						glog.Infof("DRY RUN: would send email")
						glog.V(3).Infof("%+v", email)
					} else {
						err = m.SendBounceEmail(email, item.To() /*failedRecipient*/)
						if err != nil {
							return err
						}
					}
				}

				if rule.Action.Bounce {
					email := mailgun.Email{
						From:      m.MailerDaemon(),
						To:        item.From(),
						Text:      generateBounceEmailText(item),
						Subject:   "Re: " + item.Message.Headers.Subject,
						Reference: "<" + item.Message.Headers.MessageId + ">",
					}

					if *dryRun {
						glog.Infof("DRY RUN: would send email")
						glog.V(3).Infof("%+v", email)

					} else {
						err = m.SendBounceEmail(email, item.To() /*failedRecipient*/)
						if err != nil {
							return err
						}
					}
				}

				if len(rule.Action.ForwardTo) > 0 {
					if *dryRun {
						glog.Infof("DRY RUN: would forward email.", item)
						glog.V(3).Infof("%+v", item)
					} else {
						err = m.Forward(item.Storage.Key, rule.Action.ForwardTo)
						if err != nil {
							glog.Errorf("Could not forward email: %v", err)
						}
					}
				}

				if rule.Action.StopProcessing {
					break
				}
			}
		}
		glog.V(1).Infof("Processed %s->%s, match: %v", item.From(), item.To(), matched)
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
