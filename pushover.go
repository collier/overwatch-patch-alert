package main

import (
	"net/http"
	"net/url"
)

// Pushover configuration to send messages for a specific app to a specific
// device
type Pushover struct {
	AppToken  string
	UserToken string
	Device    string
}

// pushoverURL the Pushover JSON API endpoint used to send notifications.
const pushoverURL = "https://api.pushover.net/1/messages.json"

// Notify sends a Pushover notification with the included message text.
func (p *Pushover) Notify(msg string) error {
	params := url.Values{
		"token":   {p.AppToken},
		"user":    {p.UserToken},
		"message": {msg},
		"device":  {p.Device},
	}
	_, err := http.PostForm(pushoverURL, params)
	return err
}

// NotifyWithURL sends a Pushover notification with the included message text
// and URL
func (p *Pushover) NotifyWithURL(msg string, msgURL string) error {
	params := url.Values{
		"token":   {p.AppToken},
		"user":    {p.UserToken},
		"message": {msg},
		"device":  {p.Device},
		"url":     {msgURL},
	}
	_, err := http.PostForm(pushoverURL, params)
	return err
}
