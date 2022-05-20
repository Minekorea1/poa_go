package main

import (
	"flag"
	"fmt"
	"poa/context"
	"poa/poa"
	poaUpdater "poa/updater"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	VERSION_NAME                          = "v0.2.0"
	APPLICATION_UPDATE_ADDRESS            = "github.com/Minekorea1/poa_go"
	APPLICATION_UPDATE_CHECK_INTERVAL_SEC = 3600
	MQTT_BROKER_ADDRESS                   = "192.168.50.94"
	MQTT_PORT                             = 1883
	POA_INTERVAL_SEC                      = 60
)

func ternaryOP(cond bool, valTrue, valFalse interface{}) interface{} {
	if cond {
		return valTrue
	} else {
		return valFalse
	}
}

func emptyString(str string) bool {
	return strings.TrimSpace(str) == ""
}

func Initialize() *context.Context {
	context := context.NewContext()

	context.Version = VERSION_NAME
	context.Configs.UpdateAddress = ternaryOP(emptyString(context.Configs.UpdateAddress),
		APPLICATION_UPDATE_ADDRESS, context.Configs.UpdateAddress).(string)
	context.Configs.UpdateCheckIntervalSec = ternaryOP(context.Configs.UpdateCheckIntervalSec <= 0,
		APPLICATION_UPDATE_CHECK_INTERVAL_SEC, context.Configs.UpdateCheckIntervalSec).(int)
	context.Configs.MqttBrokerAddress = ternaryOP(emptyString(context.Configs.MqttBrokerAddress),
		MQTT_BROKER_ADDRESS, context.Configs.MqttBrokerAddress).(string)
	context.Configs.MqttPort = ternaryOP(context.Configs.MqttPort <= 0,
		MQTT_PORT, context.Configs.MqttPort).(int)
	context.Configs.PoaIntervalSec = ternaryOP(context.Configs.PoaIntervalSec <= 0,
		POA_INTERVAL_SEC, context.Configs.PoaIntervalSec).(int)

	return context
}

func main() {
	versionFlag := false
	flag.BoolVar(&versionFlag, "version", false, "prints the version and exit")
	flag.Parse()

	if versionFlag {
		fmt.Println(VERSION_NAME)
		return
	}

	context := Initialize()
	// context.Configs.WriteFile("config.json")

	updater := poaUpdater.NewUpdater()
	updater.Init(context)
	updater.Start()

	poa := poa.NewPoa()
	poa.Init(context)
	poa.Start()

	// for {
	// 	time.Sleep(time.Second)
	// }

	a := app.New()
	w := a.NewWindow("Hello")

	hello := widget.NewLabel("Hello Fyne!")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
	))

	w.ShowAndRun()
}
