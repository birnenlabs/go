package gcontacts

import (
	_ "encoding/xml"
)

// Minimal set of fields that are needed by other projects.

type Feed struct {
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	Id                  string                `xml:"id"`
	Name                Name                  `xml:"name"`
	PhoneNumbers        []PhoneNumber         `xml:"phoneNumber"`
	GroupMembershipInfo []GroupMembershipInfo `xml:"groupMembershipInfo"`
	Links               []Link                `xml:"link"`
}

type Name struct {
	GivenName  string `xml:"givenName"`
	FamilyName string `xml:"familyName"`
}

type PhoneNumber struct {
	Value string `xml:",chardata"`
}

type GroupMembershipInfo struct {
	Deleted string `xml:"deleted,attr"`
	Href    string `xml:"href,attr"`
}

type Link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}
