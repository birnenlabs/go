package main

import (
	glog "birnenlabs.com/go/lib/alog"
	"birnenlabs.com/go/lib/mailgun"
	"regexp"
)

func matches(item mailgun.Item, match Match) bool {

	if setAndNotEqual(match.Event, item.Event) {
		return false
	}
	if setAndNotEqual(match.Severity, item.Severity) {
		return false
	}
	if setAndNotMatches(match.To, item.To()) {
		return false
	}
	if setAndNotMatches(match.From, item.From()) {
		return false
	}
	if setAndNotMatches(match.Headers.To, item.Message.Headers.To) {
		return false
	}
	if setAndNotMatches(match.Headers.From, item.Message.Headers.From) {
		return false
	}
	if setAndNotMatches(match.Headers.Subject, item.Message.Headers.Subject) {
		return false
	}
	if match.Attempt > 0 && item.DeliveryStatus.AttemptNo != match.Attempt {
		return false
	}
	if match.Attempt < 0 && item.DeliveryStatus.AttemptNo == -match.Attempt {
		// Let's treat less than zero as negation
		return false
	}

	return true
}

// returns true if emptyOrString is set and diffrent than toCompare
func setAndNotEqual(emptyOrString string, toCompare string) bool {
	if len(emptyOrString) > 0 {
		negateMatch := false
		if emptyOrString[0] == '-' {
			emptyOrString = emptyOrString[1:]
			negateMatch = true
		}
		if negateMatch {
			return emptyOrString == toCompare
		} else {
			return emptyOrString != toCompare
		}
	}
	return false
}

// returns true if emptyOrString is set and does not match toCompare
func setAndNotMatches(emptyOrString string, toCompare string) bool {
	glog.V(2).Infof("Comparing: %q and %q.", emptyOrString, toCompare)
	if len(emptyOrString) > 0 {
		negateMatch := false
		if emptyOrString[0] == '-' {
			emptyOrString = emptyOrString[1:]
			negateMatch = true
		}
		m, err := regexp.MatchString(emptyOrString, toCompare)
		if err != nil {
			glog.Errorf("Could not match pattern %q: %v", emptyOrString, err)
			return false
		}
		if negateMatch {
			return m
		} else {
			return !m
		}
	}
	return false
}
