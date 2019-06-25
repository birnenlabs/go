// Package icy contains icy title decoder.
package icy

import (
	"bufio"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/glog"
)

var (
	streamTitle  = []byte("StreamTitle='")
	titleTimeout = 30 * time.Minute
)

// Opens icy stream and searches for the song title. Song and title will be pushed to the titleChannel.
func Open(urlString string, titleChannel chan<- string) error {
	return OpenWithTimeout(urlString, titleChannel, time.Hour*876000)
}

// Opens icy stream and searches for the song title. Song and title will be pushed to the titleChannel.
func OpenWithTimeout(urlString string, titleChannel chan<- string, timeout time.Duration) error {
	glog.V(1).Infof("Starting stream %q...", urlString)

	client := &http.Client{}
	req, err := http.NewRequest("GET", urlString, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Icy-MetaData", "1")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	lastTitle := ""
	lastTitleTime := time.Now()
	startTime := time.Now()

	for err == nil {
		var b []byte
		b, err = reader.ReadBytes(';')
		glog.V(10).Infof("%v", b)
		t := findStreamTitle(b)
		if t != nil && *t != "" && *t != lastTitle {
			glog.V(1).Infof("New title found: %q.", *t)
			titleChannel <- *t
			lastTitle = *t
			lastTitleTime = time.Now()
		}
		if lastTitleTime.Add(titleTimeout).Before(time.Now()) {
			err = fmt.Errorf("title timeout, last title found: %v", lastTitleTime.Format("2006-01-02 15:04:05"))
		}
		if startTime.Add(timeout).Before(time.Now()) {
			err = fmt.Errorf("job timeout, last title found: %v", lastTitleTime.Format("2006-01-02 15:04:05"))
		}
	}
	return err
}

func findStreamTitle(b []byte) *string {
	for i := 0; i < len(b)-len(streamTitle); i++ {
		if streamTitle[0] == b[i] {
			if isStreamTitle(b, i) {
				// StreamTitle='Artist - Title';
				// Remove "StreamTitle='" and
				// last two characters (';)
				res := string(b[i+len(streamTitle) : len(b)-2])
				return &res
			}
		}
	}
	return nil
}

func isStreamTitle(b []byte, pos int) bool {
	for i := 0; i < len(streamTitle); i++ {
		if streamTitle[i] != b[pos+i] {
			return false
		}
	}
	return true
}
