package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	poaAutoStart "poa/autostart"
	"poa/console"
	"poa/context"
	"poa/event"
	nettool "poa/netTool"
	"poa/poa"
	"poa/res"
	"poa/ui"
	poaUpdater "poa/updater"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

const (
	VERSION_NAME                          = "v0.5.2"
	APPLICATION_UPDATE_ADDRESS            = "github.com/Minekorea1/poa_go"
	APPLICATION_UPDATE_CHECK_INTERVAL_SEC = 3600
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
	context.Configs.PoaIntervalSec = ternaryOP(context.Configs.PoaIntervalSec <= 0,
		POA_INTERVAL_SEC, context.Configs.PoaIntervalSec).(int)

	return context
}

func main() {
	versionFlag := false
	consoleFlag := false
	flag.BoolVar(&versionFlag, "version", false, "prints the version and exit")
	flag.BoolVar(&consoleFlag, "c", false, "run in console mode")
	flag.Parse()

	if consoleFlag {
		if runtime.GOOS == "windows" {
			console.AttachConsole()
		}
	}

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

	eventLooper := event.NewEventLooper()
	eventLooper.Loop()
	context.EventLooper = eventLooper

	// check update
	updater := poaUpdater.NewUpdater()
	updater.Init(context)
	updater.Start()

	// create shortcut
	autoStart := poaAutoStart.NewAutoStart()
	autoStart.DeleteShortcut()
	autoStart.CreateShortcut()

	poa := poa.NewPoa()
	poa.Init(context)
	poa.Start()

	eventLooper.RegisterEventHandler(event.MAIN, func(name event.EventName, args []interface{}) {
		fmt.Println("name:", name, args)

		switch name {
		case event.EVENT_MAIN_RESTART:
			fmt.Println("event.EVENT_MAIN_RESTART")
			poaUpdater.StartSelfProcess()
			os.Exit(0)

		case event.EVENT_MAIN_MQTT_CHANGE_USER_PASSWORD:
			fmt.Println("event.EVENT_MAIN_MQTT_CHANGE_USER_PASSWORD")
			if len(args) == 2 {
				// fmt.Println("name:", args[0])
				// fmt.Println("password:", args[1])
				context.Configs.MqttUser = args[0].(string)
				context.Configs.MqttPassword = args[1].(string)
				context.WriteConfigSync()

				poaUpdater.StartSelfProcess()
				os.Exit(0)
			}

		case event.EVENT_MAIN_FORCE_UPDATE:
			fmt.Println("event.EVENT_MAIN_FORCE_UPDATE")
			updater.Update()

		case event.EVENT_MAIN_CHANGE_UPDATE_ADDRESS:
			fmt.Println("event.EVENT_MAIN_CHANGE_UPDATE_ADDRESS")
			if len(args) == 1 {
				fmt.Println("update address:", args[0])
				context.Configs.UpdateAddress = args[0].(string)
				context.WriteConfig()

				updater.Init(context)
			}
		}
	})

	if consoleFlag {
		for {
			time.Sleep(time.Second)
		}
	} else {
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
}
