package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"poa/context"
	nettool "poa/netTool"
	"poa/poa"
	"poa/res"
	"poa/ui"
	poaUpdater "poa/updater"
	"regexp"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

const (
	VERSION_NAME                          = "v0.4.1"
	APPLICATION_UPDATE_ADDRESS            = "github.com/Minekorea1/poa_go"
	APPLICATION_UPDATE_CHECK_INTERVAL_SEC = 3600
	MQTT_BROKER_ADDRESS                   = "minekorea.asuscomm.com"
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

func getClientId() string {
	rand.Seed(time.Now().UnixNano())

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 11)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	mac := nettool.GetMacAddr()
	reg, _ := regexp.Compile(`[:]`)
	mac = reg.ReplaceAllLiteralString(mac, "")
	return fmt.Sprintf("%s%s", mac, string(b))
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

	fmt.Printf("version: %s\n", VERSION_NAME)

	context := Initialize()

	if len(context.Configs.DeviceId) != 23 {
		context.Configs.DeviceId = getClientId()
		context.WriteConfig()
	}

	updater := poaUpdater.NewUpdater()
	updater.Init(context)
	updater.Start()

	poa := poa.NewPoa()
	poa.Init(context)
	poa.Start()

	// ui
	os.Setenv("FYNE_THEME", "light") // light or dark
	a := app.NewWithID("PoaApp")
	a.SetIcon(res.IconMain)
	a.Settings().SetTheme(&ui.MyTheme{})
	a.Lifecycle().SetOnStarted(func() {
		go func() {
			for {
				ui.Status.Refresh()

				time.Sleep(time.Second)
			}
		}()
	})
	a.Lifecycle().SetOnStopped(func() {
		log.Println("Lifecycle: Stopped")
	})
	win := a.NewWindow("PoA")
	a.Settings().SetTheme(&ui.MyTheme{})
	win.SetMaster()

	ui.Init(&a, context, poa)
	uiMenu := ui.Menu{}
	subContent := container.NewMax()

	// mainContent := container.NewHSplit(uiMenu.MakeMenu(), uiStatus.GetContainer())
	// mainContent.Offset = 0.2
	mainContent := container.NewBorder(nil, nil, uiMenu.MakeMenu(subContent), nil, subContent)
	win.SetContent(mainContent)
	subContent.Objects = []fyne.CanvasObject{ui.Status.GetContent()}

	win.Resize(fyne.NewSize(640, 460))
	win.ShowAndRun()
}
