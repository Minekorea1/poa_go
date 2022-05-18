package poa

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"poa/context"
	nettool "poa/netTool"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Poa struct {
	deviceType int
	deviceName string
	deviceDesc string

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
	// poa.deviceType = context.

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
	poa.mqttOpts.SetClientID(nettool.GetMacAddr())
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

		token := poa.mqttClient.Subscribe(fmt.Sprintf("%s/%s/poa/info", nettool.GetPublicIP(), nettool.GetPrivateIP()), poa.mqttQos, nil)
		token.Wait()

		// timer start
		ticker := time.NewTicker(time.Second * time.Duration(poa.intervalSec))
		go func() {
			for range ticker.C {
				token := poa.mqttClient.Publish(fmt.Sprintf("%s/%s/poa/info", nettool.GetPublicIP(), nettool.GetPrivateIP()), poa.mqttQos, false, "hello")
				token.Wait()
			}
		}()
	}()
}
