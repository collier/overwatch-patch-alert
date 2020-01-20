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

// GetBlizzTrackVersion returns the current version number of a Blizzard game
// client for a given blizztrack game ID
func GetBlizzTrackVersion(id string) (string, error) {
	url := fmt.Sprintf("https://blizztrack.com/api/%s/info/json?mode=vers", id)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	var g Game
	err = json.Unmarshal(body, &g)
	if err != nil {
		return "", err
	}
	var v string
	for _, r := range g.Regions {
		if r.Region == "us" {
			v = r.VersionsName
			break
		}
	}
	return v, nil
}

// GetBlizzTrackPatchNotesURL returns the URL to visit in the browser to view
// patch notes for a given blizztrack game ID
func GetBlizzTrackPatchNotesURL(id string) string {
	return fmt.Sprintf("https://blizztrack.com/patch_notes/%s/latest", id)
}
