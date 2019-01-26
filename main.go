package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type build struct {
	Name     string `json:"name"`
	Patch    string `json:"patch"`
	URL      string `json:"url"`
	Selector string `json:"selector"`
}

type config struct {
	ServiceOn    bool    `json:"serviceOn"`
	FailureCount int     `json:"failureCount"`
	MaxFailures  int     `json:"maxFailures"`
	POAppToken   string  `json:"pushoverAppToken"`
	POUserToken  string  `json:"pushoverUserToken"`
	PODevice     string  `json:"pushoverDevice"`
	Builds       []build `json:"builds"`
}

func scrape(b build) (string, error) {
	doc, err := goquery.NewDocument(b.URL)
	if err != nil {
		return "", err
	}
	sel := doc.Find(b.Selector).First()
	if sel.Length() == 0 {
		msg := fmt.Sprintf("selector '%s' did not return results on %s", b.Selector, b.URL)
		return "", errors.New(msg)
	}
	return sel.Text(), nil
}

func notify(c *config, msg string, u string) error {
	params := url.Values{
		"token":   {c.POAppToken},
		"user":    {c.POUserToken},
		"message": {msg},
		"device":  {c.PODevice}}
	if u != "" {
		params.Add("url", u)
	}
	_, err := http.PostForm("https://api.pushover.net/1/messages.json", params)
	return err
}

func writeConfig(c *config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("./data.json", data, 0666)
	if err != nil {
		return err
	}
	return nil
}

func readConfig() (*config, error) {
	file, err := ioutil.ReadFile("./data.json")
	if err != nil {
		log.Printf("ERROR %v", err)
		return nil, err
	}
	var conf config
	err = json.Unmarshal(file, &conf)
	if err != nil {
		log.Printf("ERROR %v", err)
		return nil, err
	}
	return &conf, nil
}

func main() {
	start := time.Now()
	logf, err := os.OpenFile("./overwatch-patch.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logf.Close()
	log.SetOutput(logf)
	conf, err := readConfig()
	if err != nil {
		log.Printf("ERROR %v", err)
		return
	}
	if !conf.ServiceOn {
		elapsed := time.Since(start)
		log.Printf("INFO service completed, service off, completed in %s", elapsed)
		return
	}
	changed := false
	errorsOccurred := false
	for i, b := range conf.Builds {
		newPatch, err := scrape(b)
		if err != nil {
			errorsOccurred = true
			log.Printf("ERROR %v", err)
		} else if newPatch != b.Patch {
			log.Printf("INFO new patch (%s) detected for %s servers", newPatch, conf.Builds[i].Name)
			changed = true
			conf.Builds[i].Patch = newPatch
			msg := fmt.Sprintf("A new Overwatch patch has been released on the %s servers.", b.Name)
			err = notify(conf, msg, conf.Builds[i].URL)
			if err != nil {
				log.Printf("ERROR %v", err)
			}
		}
	}
	if errorsOccurred {
		conf.FailureCount = conf.FailureCount + 1
		if conf.FailureCount == conf.MaxFailures {
			conf.ServiceOn = false
			log.Printf("WARN service has been turned off due to failed scrape of patch")
			msg := fmt.Sprintf(`
				Too many consecutive errors occured while scraping 
				overwatch patches, and the service is being shut down. 
				Correct configuration and turn service back on.
			`)
			err = notify(conf, msg, "")
			if err != nil {
				log.Printf("ERROR %v", err)
			}
		} else {
			msg := "WARN The service has failed %d times consecutively. The service will be turned off after %d consecutive failures"
			log.Printf(msg, conf.FailureCount, conf.MaxFailures)
		}
		err = writeConfig(conf)
		if err != nil {
			log.Printf("ERROR %v", err)
			return
		}
		return
	}
	if conf.FailureCount > 0 && !errorsOccurred {
		conf.FailureCount = 0
		err = writeConfig(conf)
		if err != nil {
			log.Printf("ERROR %v", err)
			return
		}
		log.Printf("INFO consecutive failure count reset to 0 after successfully parsing all builds")
	}
	if changed {
		err = writeConfig(conf)
		if err != nil {
			log.Printf("ERROR %v", err)
			return
		}
		elapsed := time.Since(start)
		log.Printf("INFO service completed, changes found, completed in %s", elapsed)
	} else {
		elapsed := time.Since(start)
		log.Printf("INFO service completed, no changes found, completed in %s", elapsed)
	}
}
