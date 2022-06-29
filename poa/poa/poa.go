package poa

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"poa/context"
	"poa/event"
	"poa/log"
	nettool "poa/netTool"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var logger = log.NewLogger("PoA")

type mqttLogger struct {
	log.Logger
}

func (l mqttLogger) Println(v ...interface{}) {
	l.Log(l.Level, v...)
}

func (l mqttLogger) Printf(format string, v ...interface{}) {
	l.LogFormat(l.Level, format, v...)
}

type DeviceInfo struct {
	// Timestamp int64

	DeviceId   string
	MacAddress string
	PublicIp   string
	PrivateIp  string
	Owner      string
	OwnNumber  int
	DeviceType int
	DeviceDesc string
	Version    string
}

type Poa struct {
	deviceInfo           DeviceInfo
	MqttPublishTimestamp int64

	intervalSec   int
	brokerAddress string
	brokerPort    int

	mqttClient   mqtt.Client
	mqttOpts     *mqtt.ClientOptions
	mqttQos      byte
	mqttUser     string
	mqttPassword string

	condCh            chan int
	condMqttConnectCh chan int

	context *context.Context
}

// client to server
type Request struct {
	Type string

	Register struct {
		DeviceInfo DeviceInfo
	}
}

// server to client
type Response struct {
	Type string

	Available struct {
		OwnNumbers []int
	}
}

// server to client
type Command struct {
	Type string

	Update struct {
		ForceUpdate   bool
		UpdateAddress string
	}

	Mqtt struct {
		MqttBrokerAddress string
		MqttPort          int
		MqttUser          string
		MqttPassword      string
	}

	Restart struct {
		Restart bool
	}
}

func NewPoa() *Poa {
	poa := Poa{}
	return &poa
}

func (poa *Poa) Init(context *context.Context) {
	rand.Seed(time.Now().UnixNano())

	poa.deviceInfo.DeviceId = context.Configs.DeviceId
	poa.deviceInfo.MacAddress = nettool.GetMacAddr()
	poa.deviceInfo.Owner = context.Configs.Owner
	poa.deviceInfo.OwnNumber = context.Configs.OwnNumber
	poa.deviceInfo.DeviceType = context.Configs.DeviceType
	poa.deviceInfo.DeviceDesc = context.Configs.DeviceDesc
	poa.deviceInfo.Version = context.Version

	poa.intervalSec = context.Configs.PoaIntervalSec
	poa.brokerAddress = context.Configs.MqttBrokerAddress
	poa.brokerPort = context.Configs.MqttPort
	poa.mqttQos = 1
	poa.mqttUser = context.Configs.MqttUser
	poa.mqttPassword = context.Configs.MqttPassword

	poa.condCh = make(chan int)
	poa.condMqttConnectCh = make(chan int)

	poa.context = context

	mqtt.ERROR = mqttLogger{Logger: log.Logger{Tag: "mqtt", Timestamp: true, Level: log.Fatal}}
	mqtt.CRITICAL = mqttLogger{Logger: log.Logger{Tag: "mqtt", Timestamp: true, Level: log.Error}}
	mqtt.WARN = mqttLogger{Logger: log.Logger{Tag: "mqtt", Timestamp: true, Level: log.Warning}}
	// mqtt.DEBUG = mqttLogger{Logger: log.Logger{Tag: "mqtt", Timestamp: true, Level: log.Debug}}

	poa.mqttOpts = mqtt.NewClientOptions()
	poa.mqttOpts.AddBroker(fmt.Sprintf("tcp://%s:%d", poa.brokerAddress, poa.brokerPort))
	poa.mqttOpts.SetClientID(poa.deviceInfo.DeviceId)
	poa.mqttOpts.SetUsername(poa.mqttUser)
	poa.mqttOpts.SetPassword(poa.mqttPassword)
	poa.mqttOpts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		logger.LogfV("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})
	poa.mqttOpts.SetAutoReconnect(true)
	poa.mqttOpts.OnConnect = func(client mqtt.Client) {
		logger.LogI("MQTT connected")

		poa.condMqttConnectCh <- 0

		// run once at startup
		go func() {
			poa.condCh <- 0
		}()
	}
	poa.mqttOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		logger.LogfI("MQTT connect lost: %v", err)
	}
}

func (poa *Poa) Start() {
	logger.LogI("|", poa.deviceInfo.DeviceId)
	logger.LogI("|", poa.deviceInfo.Owner)
	logger.LogI("|", poa.deviceInfo.OwnNumber)
	logger.LogI("|", poa.deviceInfo.MacAddress)
	logger.LogI("|", poa.deviceInfo.DeviceType)
	logger.LogI("|", poa.deviceInfo.DeviceDesc)
	logger.LogI("|", poa.deviceInfo.Version)

	go func() {
		// random sleep
		time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)

		poa.mqttClient = mqtt.NewClient(poa.mqttOpts)

		if token := poa.mqttClient.Connect(); token.Wait() && token.Error() != nil {
			logger.LogE(token.Error())
			time.AfterFunc(time.Second*60, poa.Start)
			return
		}

		var publicIp string
		var privateIP string

		for publicIp == "" || privateIP == "" {
			publicIp, _ = nettool.GetPublicIP()
			privateIP = nettool.GetPrivateIP()

			if publicIp == "" || privateIP == "" {
				time.Sleep(time.Second * 10)
			}
		}

		go func() {
			for {
				<-poa.condMqttConnectCh
				poa.registerMqttSubscribe()

				if poa.deviceInfo.OwnNumber <= 0 && publicIp != "" && privateIP != "" {
					poa.deviceInfo.PublicIp = publicIp
					poa.deviceInfo.PrivateIp = privateIP

					request := Request{Type: "register"}
					request.Register.DeviceInfo = poa.deviceInfo

					doc, err := json.MarshalIndent(request, "", "    ")
					if err == nil {
						token := poa.mqttClient.Publish("mine/server/request", poa.mqttQos, false, string(doc))
						token.Wait()
					} else {
						logger.LogE(err)
					}
				}
			}
		}()

		// loop start
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(poa.intervalSec))
			go func() {
				for range ticker.C {
					poa.condCh <- 0
				}
			}()

			for {
				<-poa.condCh

				var publicIp string
				var privateIP string

				publicIp, _ = nettool.GetPublicIP()
				privateIP = nettool.GetPrivateIP()

				if poa.deviceInfo.OwnNumber <= 0 {
					if publicIp != "" && privateIP != "" {
						poa.deviceInfo.PublicIp = publicIp
						poa.deviceInfo.PrivateIp = privateIP

						request := Request{Type: "register"}
						request.Register.DeviceInfo = poa.deviceInfo

						doc, err := json.MarshalIndent(request, "", "    ")
						if err == nil {
							token := poa.mqttClient.Publish("mine/server/request", poa.mqttQos, false, string(doc))
							token.Wait()
						} else {
							logger.LogE(err)
						}
					}
				} else {
					if publicIp != "" && privateIP != "" {
						poa.deviceInfo.PublicIp = publicIp
						poa.deviceInfo.PrivateIp = privateIP
					}

					if poa.deviceInfo.PublicIp != "" && poa.deviceInfo.PrivateIp != "" {
						doc, err := json.MarshalIndent(poa.deviceInfo, "", "    ")
						if err == nil {
							token := poa.mqttClient.Publish(fmt.Sprintf("mine/%s/%s/poa/info", publicIp, poa.deviceInfo.DeviceId), poa.mqttQos, false, string(doc))
							token.Wait()

							logger.LogD("publish mqtt poa message")

							poa.MqttPublishTimestamp = time.Now().Unix()
						} else {
							logger.LogE(err)
						}
					} else {
						logger.LogW("publish failed: public ip = ", publicIp, ", private ip = ", privateIP)
					}
				}
			}
		}()
	}()
}

func (poa *Poa) registerMqttSubscribe() {
	var publicIp string
	var privateIP string

	logger.LogD("register mqtt subscribe")

	for publicIp == "" || privateIP == "" {
		publicIp, _ = nettool.GetPublicIP()
		privateIP = nettool.GetPrivateIP()

		if publicIp == "" || privateIP == "" {
			time.Sleep(time.Second * 10)
		}
	}

	token := poa.mqttClient.Subscribe(fmt.Sprintf("mine/%s/%s/poa/response/#", publicIp, poa.deviceInfo.DeviceId), poa.mqttQos,
		func(client mqtt.Client, msg mqtt.Message) {
			response := Response{}
			json.Unmarshal(msg.Payload(), &response)

			poa.processResponse(&response)
		})
	token.Wait()

	token = poa.mqttClient.Subscribe(fmt.Sprintf("mine/%s/%s/poa/command/#", publicIp, poa.deviceInfo.DeviceId), poa.mqttQos,
		func(client mqtt.Client, msg mqtt.Message) {
			command := Command{}
			json.Unmarshal(msg.Payload(), &command)

			poa.processCommand(&command)
		})
	token.Wait()
}

func (poa *Poa) ForcePublish() {
	poa.condCh <- 0
}

/*
func (poa *Poa) getRandomClientId() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 5)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return fmt.Sprintf("%s_%s", poa.deviceInfo.MacAddress, string(b))
}
*/

func (poa *Poa) processResponse(response *Response) {
	if response.Type == "available" {
		poa.deviceInfo.OwnNumber = response.Available.OwnNumbers[0]
		// poa.context.Configs.OwnNumber = poa.deviceInfo.OwnNumber
		// poa.context.WriteConfig()
		poa.WriteDeviceInfo(&poa.deviceInfo)

		logger.LogD("set OwnNumber =", poa.deviceInfo.OwnNumber)

		poa.ForcePublish()
	}
}

func (poa *Poa) processCommand(command *Command) {
	switch command.Type {
	case "update":
		if command.Update.ForceUpdate {
			poa.context.EventLooper.PushEvent(event.MAIN, event.EVENT_MAIN_FORCE_UPDATE)
		} else if command.Update.UpdateAddress != "" {
			poa.context.EventLooper.PushEvent(event.MAIN, event.EVENT_MAIN_CHANGE_UPDATE_ADDRESS, command.Update.UpdateAddress)
		}

	case "mqtt":
		if command.Mqtt.MqttUser != "" && command.Mqtt.MqttPassword != "" {
			poa.context.EventLooper.PushEvent(event.MAIN, event.EVENT_MAIN_MQTT_CHANGE_USER_PASSWORD, command.Mqtt.MqttUser, command.Mqtt.MqttPassword)
		}

	case "restart":
		if command.Restart.Restart {
			poa.context.EventLooper.PushEvent(event.MAIN, event.EVENT_MAIN_RESTART)
		}
	}

}

func (poa *Poa) GetDeviceInfo() *DeviceInfo {
	return &poa.deviceInfo
}

func (poa *Poa) WriteDeviceInfo(deviceInfo *DeviceInfo) {
	poa.context.Configs.DeviceId = deviceInfo.DeviceId
	poa.context.Configs.Owner = deviceInfo.Owner
	poa.context.Configs.OwnNumber = deviceInfo.OwnNumber
	poa.context.Configs.DeviceType = deviceInfo.DeviceType
	poa.context.Configs.DeviceDesc = deviceInfo.DeviceDesc

	poa.context.WriteConfig()

}
