package main

import (
	"birnenlabs.com/automate"
	"birnenlabs.com/conf"
	"birnenlabs.com/gcontacts"
	"context"
	"flag"
	"fmt"
	"github.com/golang/glog"
)

var configFlag = flag.String("config", "sync-contacts", "Configuration")

const appName = "Contacts sync"

func main() {
	flag.Parse()
	flag.Set("alsologtostderr", "true")
	defer glog.Flush()

	ctx := context.Background()

	// Create notifier first
	cloudMessage, err := automate.Create()
	if err != nil {
		// Not exiting here, continue without cloud notifier.
		glog.Errorf("Could not create cloud message: %v", err)
	}

	var config Config
	err = conf.LoadConfigFromJson(*configFlag, &config)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exitf("Could not read configuration: %v", err)
	}

	// First create source connector
	glog.Infof("Creating source connector: %v", config.Src.Config)
	srcContacts, err := gcontacts.NewWithCustomConfig(ctx, config.Src.Config)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exitf("Could not create source connector: %v", err)
	}

	// Then create destination connectors
	dstContacts := make([]*gcontacts.Gcontacts, len(config.Dst))
	for i := range config.Dst {
		glog.Infof("Creating destination[%v] connector: %v", i, config.Dst[i].Config)
		dstContacts[i], err = gcontacts.NewWithCustomConfig(ctx, config.Dst[i].Config)
		if err != nil {
			cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
			glog.Exitf("Could not destination connector: %v", err)
		}

	}

	// List source
	srcFeed, err := srcContacts.ListContacts(ctx)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exitf("Could not list contacts: %v", err)
	}

	srcMap, err := groupToMap(srcFeed.Entries, config.Src.Group)
	if err != nil {
		cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
		glog.Exitf("Could not list contacts: %v", err)
	}

	glog.Infof("Found %v src contacts in %q group", len(srcMap), config.Src.Group)

	// Compare with destinations
	for i := range config.Dst {
		// List destination
		dstFeed, err := dstContacts[i].ListContacts(ctx)
		if err != nil {
			cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
			glog.Exitf("Could not list contacts: %v", err)
		}

		dstMap, err := groupToMap(dstFeed.Entries, config.Dst[i].Group)
		if err != nil {
			cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
			glog.Exitf("Could not list contacts: %v", err)
		}

		glog.Infof("Found %v dst contacts in %q group", len(dstMap), config.Dst[i].Group)

		// Create toAdd list and remove valid contacts from dstMap
		toAdd := make([]string, 0)
		for k, v := range srcMap {
			existing, ok := dstMap[k]
			if ok {
				if isEqual(v, existing) {
					// If equal removing from dstMap, so it wont be deleted
					delete(dstMap, k)
				} else {
					// If not equal, adding to "toAdd"
					toAdd = append(toAdd, k)
				}
			} else {
				// Does not exist - needs to be added
				toAdd = append(toAdd, k)
			}
		}

		status := fmt.Sprintf("Adding %v and removing %v contacts for %v.", len(toAdd), len(dstMap), config.Dst[i].Config)

		// at this point dstMap contains values that are not in source or are different in the source - let's delete them:
		for k, v := range dstMap {
			status = fmt.Sprintf("%v\nD: %v", status, k)

			deleteUrl := findDeleteUrl(v)
			if len(deleteUrl) == 0 {
				cloudMessage.SendFormattedCloudMessageToDefault(appName, "Could not delete contact with empty delete url", 1)
				glog.Exitf("Delete url is empty, contact: %+v", v)
			}

			err = dstContacts[i].RemoveContact(ctx, deleteUrl)
			if err != nil {
				cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
				glog.Exitf("Could not remove contact: %v", err)
			}
		}

		// and add missing items:
		for _, k := range toAdd {
			status = fmt.Sprintf("%v\nA: %v", status, k)

			entry := srcMap[k]
			err = dstContacts[i].AddContact(ctx, createRequest(entry.Name.GivenName, entry.Name.FamilyName, config.Dst[i].Group, entry.PhoneNumbers[0].Value))
			if err != nil {
				cloudMessage.SendFormattedCloudMessageToDefault(appName, err.Error(), 1)
				glog.Exitf("Could not add contact: %v", err)
			}
		}

		glog.Infof(status)

		// Notify if there were changes
		if len(toAdd)+len(dstMap) > 0 {
			cloudMessage.SendFormattedCloudMessageToDefault(appName, status, 1)
		}
	}

	cloudMessage.SendFormattedCloudMessageToDefault(appName, "Done", 0)
}

func groupToMap(entries []gcontacts.Entry, groupId string) (map[string]gcontacts.Entry, error) {
	result := make(map[string]gcontacts.Entry)

	for _, e := range entries {
		if isInGroup(e, groupId) {
			key := e.Name.GivenName + " " + e.Name.FamilyName
			_, exists := result[key]

			if exists {
				return nil, fmt.Errorf("value %q already exists", key)
			}
			result[key] = e
		}
	}

	return result, nil
}

func isInGroup(c gcontacts.Entry, groupId string) bool {
	for _, m := range c.GroupMembershipInfo {
		if m.Href == groupId && m.Deleted == "false" {
			return true
		}
	}
	return false
}

func isEqual(c1, c2 gcontacts.Entry) bool {
	if len(c1.PhoneNumbers) != len(c2.PhoneNumbers) {
		return false
	}

	for i := 0; i < len(c1.PhoneNumbers); i++ {
		if c1.PhoneNumbers[i].Value != c2.PhoneNumbers[i].Value {
			return false
		}
	}

	return c1.Name.GivenName == c2.Name.GivenName && c1.Name.FamilyName == c2.Name.FamilyName
}

func createRequest(givenName, familyName, group, phone string) string {
	tmpl := `
		<atom:entry xmlns:atom="http://www.w3.org/2005/Atom" xmlns:gd="http://schemas.google.com/g/2005">
		<atom:category scheme="http://schemas.google.com/g/2005#kind" term="http://schemas.google.com/contact/2008#contact"/>
		<gd:name>
			<gd:givenName>%v</gd:givenName>
			<gd:familyName>%v</gd:familyName>
		</gd:name>
		<gd:groupMembershipInfo deleted="false" href="%v"/>
		<gd:phoneNumber rel="http://schemas.google.com/g/2005#work" primary="true">%v</gd:phoneNumber>
		</atom:entry>`
	return fmt.Sprintf(tmpl, givenName, familyName, group, phone)
}

func findDeleteUrl(e gcontacts.Entry) string {
	for _, l := range e.Links {
		if l.Rel == "edit" {
			return l.Href
		}
	}
	return ""
}
