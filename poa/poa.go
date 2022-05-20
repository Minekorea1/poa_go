package poa

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"poa/context"
	nettool "poa/netTool"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type DeviceInfo struct {
	// Timestamp int64

	MacAddress string
	PublicIp   string
	PrivateIp  string
	Owner      string
	OwnNumber  int
	DeviceType int
	DeviceDesc string
}

type Poa struct {
	deviceInfo DeviceInfo

	intervalSec   int
	brokerAddress string
	brokerPort    int

	mqttClient mqtt.Client
	mqttOpts   *mqtt.ClientOptions
	mqttQos    byte
}

func NewPoa() *Poa {
	poa := Poa{}
	return &poa
}

func (poa *Poa) Init(context *context.Context) {
	// TODO:
	poa.deviceInfo.MacAddress = nettool.GetMacAddr()
	poa.deviceInfo.Owner = ""
	poa.deviceInfo.OwnNumber = 0
	poa.deviceInfo.DeviceType = 0
	poa.deviceInfo.DeviceDesc = ""

	poa.intervalSec = context.Configs.PoaIntervalSec
	poa.brokerAddress = context.Configs.MqttBrokerAddress
	poa.brokerPort = context.Configs.MqttPort
	poa.mqttQos = 1

	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	poa.mqttOpts = mqtt.NewClientOptions()
	poa.mqttOpts.AddBroker(fmt.Sprintf("tcp://%s:%d", poa.brokerAddress, poa.brokerPort))
	poa.mqttOpts.SetClientID(poa.deviceInfo.MacAddress + "x") //TODO: remove x
	// poa.mqttOpts.SetUsername("emqx")
	// poa.mqttOpts.SetPassword("public")
	poa.mqttOpts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
	})
	poa.mqttOpts.SetAutoReconnect(true)
	poa.mqttOpts.OnConnect = func(client mqtt.Client) {
		fmt.Println("Connected")
	}
	poa.mqttOpts.OnConnectionLost = func(client mqtt.Client, err error) {
		fmt.Printf("Connect lost: %v", err)
	}
}

func (poa *Poa) Start() {
	go func() {
		// random sleep
		rand.Seed(time.Now().UnixNano())
		time.Sleep(time.Duration(rand.Int31n(10000)) * time.Millisecond)

		poa.mqttClient = mqtt.NewClient(poa.mqttOpts)

		if token := poa.mqttClient.Connect(); token.Wait() && token.Error() != nil {
			log.Println(token.Error())
			time.AfterFunc(time.Second*60, poa.Start)
			return
		}

		publicIp, _ := nettool.GetPublicIP()
		privateIP := nettool.GetPrivateIP()

		if publicIp == "" || privateIP == "" {
			panic("failed to obtain ip address.")
		}

		token := poa.mqttClient.Subscribe(fmt.Sprintf("mine/%s/%s/poa/#", publicIp, privateIP), poa.mqttQos, nil)
		token.Wait()

		// timer start
		ticker := time.NewTicker(time.Second * time.Duration(poa.intervalSec))
		go func() {
			for range ticker.C {
				publicIp, _ := nettool.GetPublicIP()
				privateIP := nettool.GetPrivateIP()

				if publicIp != "" && privateIP != "" {
					poa.deviceInfo.PublicIp = publicIp
					poa.deviceInfo.PrivateIp = privateIP

					doc, err := json.MarshalIndent(poa.deviceInfo, "", "    ")
					if err == nil {
						token := poa.mqttClient.Publish(fmt.Sprintf("mine/%s/%s/poa/info", publicIp, privateIP), poa.mqttQos, false, string(doc))
						token.Wait()
					}
				}
			}
		}()
	}()
}
