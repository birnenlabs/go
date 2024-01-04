package main

import (
	"bytes"
	"strconv"
)

type Config struct {
	Rules []Rule
}

type Rule struct {
	Match  Match
	Action Action
}

// When more fields are set they are joind with "AND" operator
type Match struct {
	// failed, rejected, delivered etc. - the same as mailgun api
	Event string
	// temporary, permanent - currently set for failed events only
	Severity string
	// regexp to match email address of message recipient
	To string
	// regexp to match email address of message sender
	From string
	// asttempt number from the event
	Attempt int
	Headers Headers
}

type Headers struct {
	// Note: To and From are different than Match.To/From as it might contain additional information set by the sender, e.g:
	// Match.To = "me@example.com"
	// Match.Header.To == "'Somebodys Name' <me@example.com>"

	// regexp to match To header
	To string
	// regexp to match From header
	From string
	// regexp to matchsubject header
	Subject string
}

type Action struct {
	NotifyPostmaster bool
	Bounce           bool
	StopProcessing   bool
	ForwardTo        string
}

func (m Match) String() string {
	var b bytes.Buffer
	b.WriteString("{")
	if len(m.Event) > 0 {
		b.WriteString("Event: '")
		b.WriteString(m.Event)
		b.WriteString("' ")
	}
	if len(m.Severity) > 0 {
		b.WriteString("Severity: '")
		b.WriteString(m.Severity)
		b.WriteString("' ")
	}
	if len(m.To) > 0 {
		b.WriteString("To: '")
		b.WriteString(m.To)
		b.WriteString("' ")
	}
	if len(m.From) > 0 {
		b.WriteString("From: '")
		b.WriteString(m.From)
		b.WriteString("' ")
	}
	if m.Attempt != 0 {
		b.WriteString("Attempt: '")
		b.WriteString(strconv.Itoa(m.Attempt))
		b.WriteString("' ")
	}
	b.WriteString("Headers: ")
	b.WriteString(m.Headers.String())

	b.WriteString("}")
	return b.String()
}

func (h Headers) String() string {
	var b bytes.Buffer
	b.WriteString("{")
	if len(h.To) > 0 {
		b.WriteString("To: '")
		b.WriteString(h.To)
		b.WriteString("' ")
	}
	if len(h.From) > 0 {
		b.WriteString("From: '")
		b.WriteString(h.From)
		b.WriteString("' ")
	}
	if len(h.Subject) > 0 {
		b.WriteString("Subject: '")
		b.WriteString(h.Subject)
		b.WriteString("' ")
	}

	b.WriteString("}")
	return b.String()

}

func (a Action) String() string {
	var b bytes.Buffer
	b.WriteString("{")
	if a.NotifyPostmaster {
		b.WriteString("NotifyPostmaster ")
	}
	if a.Bounce {
		b.WriteString("Bounce ")
	}
	if a.StopProcessing {
		b.WriteString("StopProcessing")
	}
	if len(a.ForwardTo) > 0 {
		b.WriteString("ForwardTo: '")
		b.WriteString(a.ForwardTo)
		b.WriteString("' ")
	}
	b.WriteString("}")
	return b.String()
}
