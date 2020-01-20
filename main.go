package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gregdel/pushover"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var wg sync.WaitGroup

func init() {
	log.SetLevel(log.InfoLevel)

	viper.SetConfigFile("config.toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Panic("unable to read config file")
	}
}

func main() {
	start := time.Now()
	logf := initLogFile()
	defer logf.Close()

	// exit when service is not active
	active := viper.GetBool("active")
	if !active {
		log.Info("service is not active")
		return
	}

	// do Live and PTR checks concurrently
	results := make(chan *versionCheckResult, 2)
	wg.Add(2)
	go doVersionCheck("overwatch", results)
	go doVersionCheck("overwatch_ptr", results)
	wg.Wait()
	close(results)
	hadNew := false
	hadErrors := false
	for r := range results {
		if r.Errored {
			hadErrors = true
		}
		if r.Version != "" {
			hadNew = true
			log.Infof("new patch v%s for %s", r.Version, r.ID)
			viper.Set(r.ID+".version", r.Version)
		}
	}

	// turn off service if failures exceed maximum allowed by configuration
	failCount := viper.GetInt("failures")
	maxFails := viper.GetInt("maxFailures")
	if hadErrors {
		failCount++
		viper.Set("failures", failCount)
		log.Warnf("service has failed %d times, and deactivates after %d failures", failCount, maxFails)
	} else if failCount > 0 {
		viper.Set("failures", 0)
		log.Info("failure count has been reset after successfully fetching patches")
	}
	if failCount >= maxFails {
		viper.Set("active", false)
		// reset failure count after deactivating service
		viper.Set("failures", 0)
		log.Warn("service deactived, max allowed failures exceeded")
		sendMessage(&pushover.Message{
			Title:   "Service Deactivated",
			Message: "Too many errors occured while checking for Overwatch patches",
		})
	}

	// write to config file if new patch found or if failures occured
	if hadNew || hadErrors {
		err := viper.WriteConfig()
		if err != nil {
			log.Error("unable to write config file")
			return
		}
	}

	elapsed := time.Since(start)
	log.Infof("service completed successfully in %s", elapsed)
}

func initLogFile() *os.File {
	fp := "./overwatch-patch-alert.log"
	f, err := os.OpenFile(fp, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panic("unable to read log file")
	}
	log.SetOutput(f)
	return f
}

type versionCheckResult struct {
	ID      string
	Version string
	Errored bool
}

func doVersionCheck(id string, results chan *versionCheckResult) {
	defer wg.Done()
	game := viper.Sub(id)
	name := game.GetString("name")
	v0 := game.GetString("version")
	v1, err := GetBlizzTrackVersion(id)
	if err != nil {
		log.Error("failed to get new version from BlizzTrack API")
		results <- &versionCheckResult{
			ID:      id,
			Errored: true,
		}
		return
	}
	if v0 != v1 {
		// send pushover notification about new patch
		title := fmt.Sprintf("New %s Overwatch Patch", name)
		msg := fmt.Sprintf("A new Overwatch patch has been released on the %s servers.", name)
		url := GetBlizzTrackPatchNotesURL(id)
		sendMessage(&pushover.Message{
			Title:   title,
			Message: msg,
			URL:     url,
		})
		results <- &versionCheckResult{
			ID:      id,
			Version: v1,
		}
		return
	}
	results <- &versionCheckResult{
		ID: id,
	}
}

func sendMessage(m *pushover.Message) {
	po := viper.Sub("pushover")
	appt := po.GetString("appToken")
	app := pushover.New(appt)
	rt := po.GetString("userToken")
	r := pushover.NewRecipient(rt)
	d := po.GetString("device")

	// modify pushover message to set the device from config
	m.DeviceName = d

	_, err := app.SendMessage(m, r)
	if err != nil {
		log.Error("failed to send pushover notification")
	}
}
