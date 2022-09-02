package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	poaAutoStart "poa/autostart"
	"poa/console"
	poaContext "poa/context"
	"poa/event"
	"poa/log"
	nettool "poa/netTool"
	"poa/poa"
	"poa/server"
	poaUpdater "poa/updater"
	"watcher/watcher"

	"google.golang.org/grpc"
)

const (
	VERSION_NAME                          = "v0.5.7"
	APPLICATION_UPDATE_ADDRESS            = "github.com/Minekorea1/poa_go"
	APPLICATION_UPDATE_CHECK_INTERVAL_SEC = 3600
	POA_INTERVAL_SEC                      = 60
	GRPC_PORT                             = "8889"
)

var logger = log.NewLogger("PoA main")

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

func Initialize() *poaContext.Context {
	context := poaContext.NewContext()

	context.Version = VERSION_NAME
	context.Configs.UpdateAddress = ternaryOP(emptyString(context.Configs.UpdateAddress),
		APPLICATION_UPDATE_ADDRESS, context.Configs.UpdateAddress).(string)
	context.Configs.UpdateCheckIntervalSec = ternaryOP(context.Configs.UpdateCheckIntervalSec <= 0,
		APPLICATION_UPDATE_CHECK_INTERVAL_SEC, context.Configs.UpdateCheckIntervalSec).(int)
	context.Configs.PoaIntervalSec = ternaryOP(context.Configs.PoaIntervalSec <= 0,
		POA_INTERVAL_SEC, context.Configs.PoaIntervalSec).(int)

	return context
}

func runRpcClient(poa *poa.Poa) {
	for {
		conn, err := grpc.Dial("localhost:"+GRPC_PORT, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			logger.LogfE("did not connect: %v", err)
		}
		// defer conn.Close()

		c := watcher.NewWatcherServiceClient(conn)

		// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		// defer cancel()

		void, err := c.ImAlive(context.Background(), &watcher.AliveInfo{MqttPublishTimestamp: poa.MqttPublishTimestamp, PoaArgs: os.Args})
		_ = void
		if err != nil {
			logger.LogfE("could not request: %v", err)
		}

		time.Sleep(time.Second * 1)
	}
}

func runWebGuiAndWait(poaCtx *poaContext.Context, poa *poa.Poa) {
	poaHttpServer := server.NewHttpServer(poaCtx, poa)

	go func() {
		webContext := poaHttpServer.GetWebContext()
		webContext.MqttBrokerAddress = poaCtx.Configs.MqttBrokerAddress
		webContext.MqttPort = strconv.Itoa(poaCtx.Configs.MqttPort)
		webContext.MqttUser = poaCtx.Configs.MqttUser
		webContext.MqttPassword = poaCtx.Configs.MqttPassword

		for {
			webContext.DeviceId = poa.GetDeviceInfo().DeviceId
			webContext.MacAddress = poa.GetDeviceInfo().MacAddress
			webContext.PublicIp = poa.GetDeviceInfo().PublicIp
			webContext.PrivateIp = poa.GetDeviceInfo().PrivateIp
			webContext.Owner = poa.GetDeviceInfo().Owner
			webContext.OwnNumber = strconv.Itoa(poa.GetDeviceInfo().OwnNumber)
			webContext.DeviceType = strconv.Itoa(poa.GetDeviceInfo().DeviceType)
			webContext.DeviceDesc = poa.GetDeviceInfo().DeviceDesc
			webContext.Version = poa.GetDeviceInfo().Version
			webContext.MqttPublishTimestamp = time.Unix(poa.MqttPublishTimestamp, 0).Format("2006-01-02 15:04:05")

			time.Sleep(time.Second * 3)
		}
	}()

	go func() {
		time.Sleep(time.Second)
		if _, err := poaUpdater.StartProcess("google-chrome", "--window-size=600,900", "http://127.0.0.1:8000"); err != nil {
			logger.LogE(err)

			if _, err = poaUpdater.StartProcess("explorer", "http://127.0.0.1:8000"); err != nil {
				logger.LogE(err)
			}
		}
	}()

	if err := http.ListenAndServe(":8000", poaHttpServer.NewHttpHandler()); err != nil {
		logger.LogE(err)
	}
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

	logger.LogfI("version: %s\n", VERSION_NAME)

	poaContext := Initialize()

	if len(poaContext.Configs.DeviceId) != 23 {
		poaContext.Configs.DeviceId = getClientId()
		poaContext.WriteConfig()
	}

	eventLooper := event.NewEventLooper()
	eventLooper.Loop()
	poaContext.EventLooper = eventLooper

	// check update
	updater := poaUpdater.NewUpdater()
	updater.Init(poaContext)
	updater.Start()

	// create shortcut
	autoStart := poaAutoStart.NewAutoStart()
	autoStart.DeleteShortcut()
	autoStart.CreateShortcut()

	poa := poa.NewPoa()
	poa.Init(poaContext)
	poa.Start()

	eventLooper.RegisterEventHandler(event.MAIN, func(name event.EventName, args []interface{}) {
		logger.LogD("name:", name, args)

		switch name {
		case event.EVENT_MAIN_RESTART:
			logger.LogD("event.EVENT_MAIN_RESTART")
			poaUpdater.StartSelfProcess()
			os.Exit(0)

		case event.EVENT_MAIN_MQTT_CHANGE_USER_PASSWORD:
			logger.LogD("event.EVENT_MAIN_MQTT_CHANGE_USER_PASSWORD")
			if len(args) == 2 {
				// logger.LogD("name:", args[0])
				// logger.LogD("password:", args[1])
				poaContext.Configs.MqttUser = args[0].(string)
				poaContext.Configs.MqttPassword = args[1].(string)
				poaContext.WriteConfigSync()

				poaUpdater.StartSelfProcess()
				os.Exit(0)
			}

		case event.EVENT_MAIN_FORCE_UPDATE:
			logger.LogD("event.EVENT_MAIN_FORCE_UPDATE")
			updater.Update()

		case event.EVENT_MAIN_CHANGE_UPDATE_ADDRESS:
			logger.LogD("event.EVENT_MAIN_CHANGE_UPDATE_ADDRESS")
			if len(args) == 1 {
				logger.LogD("update address:", args[0])
				poaContext.Configs.UpdateAddress = args[0].(string)
				poaContext.WriteConfig()

				updater.Init(poaContext)
			}
		}
	})

	go runRpcClient(poa)

	runWebGuiAndWait(poaContext, poa)

	/*
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
				logger.LogD("Lifecycle: Stopped")
			})
			win := a.NewWindow("PoA")
			a.Settings().SetTheme(&ui.MyTheme{})
			win.SetMaster()

			ui.Init(&a, &win, poaContext, poa)
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
	*/
}
