package automate

import (
	"birnenlabs.com/conf"
	"bytes"
	"fmt"
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
	Secret string
}

func Create() (*CloudMessage, error) {
	var result CloudMessage
	err := conf.LoadConfigFromJson("cloud-message", &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *CloudMessage) SendCloudMessage(to string, payload string) error {

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

func (c *CloudMessage) SendFormattedCloudMessage(from string, to string, msg string, priority int) error {
	hostname, _ := os.Hostname()
	from = strings.Replace(from, "|", "", -1)
	msg = strings.Replace(msg, "|", "", -1)
	payload := fmt.Sprintf("%v [%v]|%v|%v", from, hostname, priority, msg)
	return c.SendCloudMessage(to, payload)
}
