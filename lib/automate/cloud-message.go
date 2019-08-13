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
	allowedCharacters = `[^a-zA-Z0-9 ~!@#$%^&*()_+=\[\]{}"'|\\-]+`
)

type CloudMessage struct {
	Secret    string
	DefaultTo string
}

func Create() (*CloudMessage, error) {
	var result CloudMessage
	err := conf.LoadConfigFromJson("cloud-message", &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// All the methods of CloudMessage should support nil pointer!

func (c *CloudMessage) SendCloudMessageToDefault(payload string) error {
	if c == nil {
		glog.Errorf("Could not send: %q.", payload)
		return fmt.Errorf("%q not sent", payload)
	}

	return c.SendCloudMessage(c.DefaultTo, payload)
}

func (c *CloudMessage) SendCloudMessage(to string, payload string) error {
	if c == nil {
		glog.Errorf("Could not send: %q.", payload)
		return fmt.Errorf("%q not sent", payload)
	}

	// Replace newlines and remove non printable characters
	payload = strings.Replace(payload, "\n", `\n`, -1)
	reg, err := regexp.Compile(allowedCharacters)
	if err != nil {
		return err
	}
	payload = reg.ReplaceAllString(payload, " ")

	var jsonStr = []byte(fmt.Sprintf(payloadFormat, c.Secret, to, payload))
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return nil
	} else {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf(string(body))
	}
}

func (c *CloudMessage) SendFormattedCloudMessageToDefault(from string, msg string, priority int) error {
	return c.SendCloudMessageToDefault(createFormattedPayload(from, msg, priority))
}

func (c *CloudMessage) SendFormattedCloudMessage(from string, to string, msg string, priority int) error {
	return c.SendCloudMessage(to, createFormattedPayload(from, msg, priority))
}

func createFormattedPayload(from string, msg string, priority int) string {
	hostname, _ := os.Hostname()
	from = strings.Replace(from, "|", "", -1)
	msg = strings.Replace(msg, "|", "", -1)
	return fmt.Sprintf("%v [%v]|%v|%v", from, hostname, priority, msg)
}
