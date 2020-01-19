package main

import (
	"fmt"
	"os"
	"sync"

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
	f := initLogFile()
	defer f.Close()

	// exit when service is not active
	active := viper.GetBool("active")
	if !active {
		log.Info("service is not active")
		return
	}

	wg.Add(2)
	liveV, _ := doVersionCheck("live")
	fmt.Println(*liveV)
}

func initLogFile() *os.File {
	n := "./overwatch-patch-alert.log"
	f, err := os.OpenFile(n, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Panic("unable to read log file")
	}
	log.SetOutput(f)
	return f
}

func doVersionCheck(name string) (*string, error) {
	defer wg.Done()
	id := viper.GetString(name + ".id")
	vkey := name + ".version"
	v0 := viper.GetString(vkey)
	v1, err := GetBlizzTrackVersion(id)
	if err != nil {
		return nil, err
	}
	if v0 != v1 {
		log.Infof("new patch v%s for %s", id, v1)
		viper.Set(vkey, v1)
		url := GetBlizzTrackPatchNotesURL(id)
		msg := fmt.Sprintf("A new Overwatch patch has been released on the %s servers.", name)
		sendMessage(&pushover.Message{
			Message: msg,
			URL:     url,
		})
		return &v1, nil
	}
	return nil, nil
}

func sendMessage(m *pushover.Message) {
	appt := viper.GetString("pushover.appToken")
	app := pushover.New(appt)

	rt := viper.GetString("pushover.userToken")
	r := pushover.NewRecipient(rt)

	d := viper.GetString("pushover.device")
	m.DeviceName = d

	_, err := app.SendMessage(m, r)
	if err != nil {
		log.Error("failed to send pushover notification")
	}
}
