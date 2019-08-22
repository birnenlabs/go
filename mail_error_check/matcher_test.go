package main

import (
	"birnenlabs.com/lib/mailgun"
	"testing"
)

type regexpAndString struct {
	regexp string
	value  string
	result bool
}

type matchAndItem struct {
	match Match
	item  mailgun.Item
}

func TestMatches(t *testing.T) {
	for _, s := range []matchAndItem{
		{
			match: Match{},
			item:  mailgun.Item{},
		},
		{
			match: Match{
				Event: "event",
			},
			item: mailgun.Item{
				Event: "event",
			},
		},
		{
			match: Match{
				To: "to",
			},
			item: mailgun.Item{
				Envelope: mailgun.Envelope{
					Targets: "to",
				},
			},
		},
		{
			match: Match{
				To: "to",
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						To: "Someone <to>",
					},
				},
			},
		},
		{
			match: Match{
				From: "from",
			},
			item: mailgun.Item{
				Envelope: mailgun.Envelope{
					Sender: "from",
				},
			},
		},
		{
			match: Match{
				From: "from",
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From: "Someone <from>",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					To: "to",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						To: "to",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					From: "from",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From: "from",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					Subject: "subject",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						Subject: "subject",
					},
				},
			},
		},
		{
			match: Match{
				Event: "event",
				To:    "to",
				From:  "from",
				Headers: Headers{
					From:    "Me <from>",
					To:      "You <to>",
					Subject: "subject",
				},
			},
			item: mailgun.Item{
				Event: "event",
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From:    "Me <from>",
						To:      "You <to>",
						Subject: "subject",
					},
				},
			},
		},
		{
			match: Match{
				Event: "-event2",
			},
			item: mailgun.Item{
				Event: "event",
			},
		},
		{
			match: Match{
				To: "-to2",
			},
			item: mailgun.Item{
				Envelope: mailgun.Envelope{
					Targets: "to",
				},
			},
		},
		{
			match: Match{
				From: "-from2",
			},
			item: mailgun.Item{
				Envelope: mailgun.Envelope{
					Sender: "from",
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					To: "-to2",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						To: "to",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					From: "-from2",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From: "from",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					Subject: "-subject2",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						Subject: "subject",
					},
				},
			},
		},
	} {
		if !matches(s.item, s.match) {
			t.Errorf("failed item: %+v, match: %+v ", s.item, s.match)
		}
	}
}

func TestDoesntMatch(t *testing.T) {
	for _, s := range []matchAndItem{
		{
			match: Match{
				Event: "event",
			},
			item: mailgun.Item{},
		},
		{
			match: Match{
				To: "to",
			},
			item: mailgun.Item{},
		},
		{
			match: Match{
				From: "from",
			},
			item: mailgun.Item{},
		},
		{
			match: Match{
				Headers: Headers{
					To: "to",
				},
			},
			item: mailgun.Item{},
		},
		{
			match: Match{
				Headers: Headers{
					From: "from",
				},
			},
			item: mailgun.Item{},
		},
		{
			match: Match{
				Headers: Headers{
					Subject: "subject",
				},
			},
			item: mailgun.Item{},
		},
		{
			match: Match{
				Event: "event",
				To:    "to",
				From:  "from",
				Headers: Headers{
					From:    "Me <from>",
					To:      "You <to>",
					Subject: "subject",
				},
			},
			item: mailgun.Item{
				Event: "event",
				Envelope: mailgun.Envelope{
					Targets: "override header",
				},
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From:    "Me <from>",
						To:      "You <to>",
						Subject: "subject",
					},
				},
			},
		},
		{
			match: Match{
				Event: "event",
				To:    "to",
				From:  "from",
				Headers: Headers{
					From:    "Me <from>",
					To:      "You <to>",
					Subject: "subject",
				},
			},
			item: mailgun.Item{
				Event: "event",
				Envelope: mailgun.Envelope{
					Sender: "override header",
				},
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From:    "Me <from>",
						To:      "You <to>",
						Subject: "subject",
					},
				},
			},
		},
		{
			match: Match{
				Event: "-event",
			},
			item: mailgun.Item{
				Event: "event",
			},
		},
		{
			match: Match{
				To: "-to",
			},
			item: mailgun.Item{
				Envelope: mailgun.Envelope{
					Targets: "to",
				},
			},
		},
		{
			match: Match{
				To: "-to",
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						To: "Someone <to>",
					},
				},
			},
		},
		{
			match: Match{
				From: "-from",
			},
			item: mailgun.Item{
				Envelope: mailgun.Envelope{
					Sender: "from",
				},
			},
		},
		{
			match: Match{
				From: "-from",
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From: "Someone <from>",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					To: "-to",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						To: "to",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					From: "-from",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						From: "from",
					},
				},
			},
		},
		{
			match: Match{
				Headers: Headers{
					Subject: "-subject",
				},
			},
			item: mailgun.Item{
				Message: mailgun.Message{
					Headers: mailgun.Headers{
						Subject: "subject",
					},
				},
			},
		},
	} {
		if matches(s.item, s.match) {
			t.Errorf("failed item: %+v, match: %+v ", s.item, s.match)
		}
	}
}

func TestAllRegexp(t *testing.T) {

	for _, r := range []regexpAndString{
		{
			// if regexp not set should be treated as empty
			regexp: "",
			value:  "a",
			result: true,
		},
		{
			regexp: "^$",
			value:  "",
			result: true,
		},
		{
			regexp: "^$",
			value:  "a",
			result: false,
		},
		{
			regexp: "^.$",
			value:  "d",
			result: true,
		},
		{
			regexp: "^.$",
			value:  "dd",
			result: false,
		},
		{
			regexp: "part",
			value:  "this is part of bigger",
			result: true,
		},
		{
			regexp: "^prefix",
			value:  "prefix of something bigger",
			result: true,
		},
		{
			regexp: "suffix$",
			value:  "something with suffix",
			result: true,
		},
		{
			regexp: "-^$",
			value:  "",
			result: false,
		},
		{
			regexp: "-^$",
			value:  "a",
			result: true,
		},
		{
			regexp: "-^.$",
			value:  "d",
			result: false,
		},
		{
			regexp: "-^.$",
			value:  "dd",
			result: true,
		},
		{
			regexp: "-part",
			value:  "this is part of bigger",
			result: false,
		},
		{
			regexp: "-^prefix",
			value:  "prefix of something bigger",
			result: false,
		},
		{
			regexp: "-suffix$",
			value:  "something with suffix",
			result: false,
		},
	} {
		for _, s := range []matchAndItem{
			{
				match: Match{
					To: r.regexp,
				},
				item: mailgun.Item{
					Envelope: mailgun.Envelope{
						Targets: r.value,
					},
				},
			},
			{
				match: Match{
					From: r.regexp,
				},
				item: mailgun.Item{
					Envelope: mailgun.Envelope{
						Sender: r.value,
					},
				},
			},
			{
				match: Match{
					Headers: Headers{
						To: r.regexp,
					},
				},
				item: mailgun.Item{
					Message: mailgun.Message{
						Headers: mailgun.Headers{
							To: r.value,
						},
					},
				},
			},
			{
				match: Match{
					Headers: Headers{
						From: r.regexp,
					},
				},
				item: mailgun.Item{
					Message: mailgun.Message{
						Headers: mailgun.Headers{
							From: r.value,
						},
					},
				},
			},
			{
				match: Match{
					Headers: Headers{
						Subject: r.regexp,
					},
				},
				item: mailgun.Item{
					Message: mailgun.Message{
						Headers: mailgun.Headers{
							Subject: r.value,
						},
					},
				},
			},
		} {
			got := matches(s.item, s.match)
			if got != r.result {
				t.Errorf("got: %v, want: %v, failed item: %+v, match: %+v ", got, r.result, s.item, s.match)
			}

		}
	}

}
