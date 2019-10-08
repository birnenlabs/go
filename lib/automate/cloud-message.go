package automate

import (
	"birnenlabs.com/lib/conf"
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	payloadFormat = `
		{
		"secret":   "%v",
		"to":       "%v",
		"device":   null,
		"priority": "normal",
		"payload":  "%v"
		}`
	url               = "https://llamalab.com/automate/cloud/message"
	allowedCharacters = `[^a-zA-Z0-9 ~!@#$%^&*()_+=\[\]{}|\\-]+`
	pUrgent           = 1
	pNormal           = 0
)

type CloudMessage struct {
	// Secret and DefaultTo can be stored in config. Secret is required.
	// https://llamalab.com/automate/cloud/ - generate secret.
	Secret    string
	DefaultTo string

	// appName can be set to use formatted version
	appName string
}

func Create() (*CloudMessage, error) {
	var result CloudMessage
	err := conf.LoadConfigFromJson("cloud-message", &result)
	if err != nil {
		return nil, err
	}

	if len(result.Secret) == 0 {
		return nil, fmt.Errorf("Secret has to be set.")
	}

	return &result, nil
}

// When appName is set, the formatted payload will be used:
// $from [$hostname]|$priority|$message
func (c *CloudMessage) UseFormattedPayload(appName string) {
	if c != nil {
		c.appName = appName
	}
}

// All the methods of CloudMessage should support nil pointer!

// Sends message to the default recipient.
func (c *CloudMessage) Send(msg string) {
	c.send(pNormal, "", msg)
}

// Sends message to the default recipient with priority set to urgent.
// Note: Formatted payload with priority is supported only if UseFormattedPayload was called.
func (c *CloudMessage) SendUrgent(msg string) {
	c.send(pUrgent, "", msg)
}

// Sends message to the custom recipient.
func (c *CloudMessage) SendTo(to string, msg string) {
	c.send(pNormal, to, msg)
}

// Sends message to the custom recipient with priority set to urgent.
// Note: Formatted payload with priority is supported only if UseFormattedPayload was called.
func (c *CloudMessage) SendUrgentTo(to string, msg string) {
	c.send(pUrgent, to, msg)
}

func (c *CloudMessage) send(priority int, to string, msg string) {
	if c == nil {
		glog.Errorf("Cloud message not initialized! NOT SENT: %q.", msg)
		return
	}

	// Use DefaultTo if not set.
	if len(to) == 0 {
		to = c.DefaultTo
	}

	// Use formatted payload when appName is set.
	if len(c.appName) > 0 {
		msg = createFormattedPayload(c.appName, msg, priority)
	}

	// Replace newlines and remove non printable characters
	payload := strings.Replace(msg, "\n", `\n`, -1)
	reg, err := regexp.Compile(allowedCharacters)
	if err != nil {
		glog.Errorf("Could not send cloud message %q: %v.", msg, err)
		return
	}
	payload = reg.ReplaceAllString(payload, " ")

	var jsonStr = []byte(fmt.Sprintf(payloadFormat, c.Secret, to, payload))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("Could not send cloud message %q: %v.", msg, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := ioutil.ReadAll(resp.Body)
		glog.Errorf("Could not send cloud message %q, response:\n%v", msg, string(body))
	}
}

func createFormattedPayload(from string, msg string, priority int) string {
	hostname, _ := os.Hostname()
	from = strings.Replace(from, "|", "", -1)
	msg = strings.Replace(msg, "|", "", -1)
	return fmt.Sprintf("%v [%v]|%v|%v", from, hostname, priority, msg)
}
