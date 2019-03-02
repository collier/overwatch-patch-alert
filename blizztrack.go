package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Region contains the client version details for each player region
type Region struct {
	BuildConfig   string `json:"buildconfig"`
	BuildID       string `json:"buildid"`
	CdnConfig     string `json:"cdnconfig"`
	Keyring       string `json:"keyring"`
	Region        string `json:"region"`
	RegionName    string `json:"regionname"`
	VersionsName  string `json:"versionsname"`
	ProductConfig string `json:"productconfig"`
	Updated       string `json:"updated"`
}

// Game contains the details about the Blizzard game returned by the BlizzTrack
// API
type Game struct {
	Name      string   `json:"name"`
	URL       string   `json:"url"`
	GameType  string   `json:"game_type"`
	NotesCode string   `json:"notes_code"`
	Code      string   `json:"code"`
	UseTact   bool     `json:"use_tact"`
	Regions   []Region `json:"regions"`
}

// GetCurrentVersion Returns the current version number of a Blizzard game
// client in the US region
func GetCurrentVersion(blizztrackID string) (string, error) {
	urlTmpl := "https://blizztrack.com/api/%s/info/json?mode=vers"
	url := fmt.Sprintf(urlTmpl, blizztrackID)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var game Game
	err = json.Unmarshal(body, &game)
	if err != nil {
		return "", err
	}
	var version string
	for _, r := range game.Regions {
		if r.Region == "US" {
			version = r.VersionsName
		}
	}
	return version, nil
}

// GetPatchNotesURL returns the URL to visit in the browser to view patch notes
// for a given blizztrackID
func GetPatchNotesURL(blizztrackID string) string {
	urlTmpl := "https://blizztrack.com/patch_notes/%s/latest"
	return fmt.Sprintf(urlTmpl, blizztrackID)
}
