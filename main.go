package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	start := time.Now()
	logf, err := os.OpenFile("./overwatch-patch-alert.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logf.Close()
	log.SetOutput(logf)

	conf, err := GetConfig()
	if err != nil {
		log.Printf("ERROR %v", err)
		return
	}

	pushover := &Pushover{
		AppToken:  conf.POAppToken,
		UserToken: conf.POUserToken,
		Device:    conf.PODevice,
	}

	if !conf.ServiceOn {
		elapsed := time.Since(start)
		log.Printf("INFO service off, completed in %s", elapsed)
		return
	}

	newVersionFound := false
	errorsOccurred := false
	for i, gc := range conf.GameClients {
		newVersion, err := GetCurrentVersion(gc.BlizztrackID)
		if err != nil {
			errorsOccurred = true
			log.Printf("ERROR %v", err)
		} else if newVersion == "" {
			errorsOccurred = true
			log.Printf("ERROR failed to get new version from blizztrack API")
		} else if newVersion != gc.Version {
			log.Printf("INFO new patch (%s) for %s servers", newVersion, gc.Name)
			newVersionFound = true
			conf.GameClients[i].Version = newVersion
			msg := fmt.Sprintf("A new Overwatch patch has been released on the %s servers.", gc.Name)
			msgURL := GetPatchNotesURL(gc.BlizztrackID)
			err = pushover.NotifyWithURL(msg, msgURL)
			if err != nil {
				log.Printf("ERROR failed to send pushover notification: %v", err)
			}
		}
	}

	if errorsOccurred {
		conf.FailureCount = conf.FailureCount + 1
		if conf.FailureCount >= conf.MaxFailures {
			conf.ServiceOn = false
			log.Printf("WARN service has been turned off due to failed scrape of patch")
			msg := fmt.Sprintf("Too many consecutive errors occured while scraping overwatch patches, and the service is being shut down. Correct configuration and turn service back on.")
			err = pushover.Notify(msg)
			if err != nil {
				log.Printf("ERROR failed to send pushover notification: %v", err)
			}
		} else {
			msg := "WARN service has failed %d times consecutively, and will be turned off after %d consecutive failures"
			log.Printf(msg, conf.FailureCount, conf.MaxFailures)
		}
	}

	failuresReset := false
	if conf.FailureCount > 0 && !errorsOccurred {
		failuresReset = true
		conf.FailureCount = 0
		log.Printf("INFO consecutive failure count reset to 0 after successfully parsing all builds")
	}

	if newVersionFound || errorsOccurred || failuresReset {
		err = conf.WriteToFile()
		if err != nil {
			log.Printf("ERROR unable to write to config.json: %v", err)
			return
		}
	}

	elapsed := time.Since(start)
	log.Printf("INFO service completed successfully in %s", elapsed)
}
