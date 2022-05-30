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

	mqttClient mqtt.Client
	mqttOpts   *mqtt.ClientOptions
	mqttQos    byte

	condCh chan int

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
		Update bool
	}

	Address struct {
		UpdateAddress     string
		MqttBrokerAddress string
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

	poa.condCh = make(chan int)

	poa.context = context

	mqtt.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
	mqtt.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
	mqtt.WARN = log.New(os.Stdout, "[WARN]  ", 0)
	// mqtt.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)

	poa.mqttOpts = mqtt.NewClientOptions()
	poa.mqttOpts.AddBroker(fmt.Sprintf("tcp://%s:%d", poa.brokerAddress, poa.brokerPort))
	poa.mqttOpts.SetClientID(poa.deviceInfo.DeviceId)
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

				fmt.Println("_____ cmd:", command)
			})
		token.Wait()

		if poa.deviceInfo.OwnNumber <= 0 {
			poa.deviceInfo.PublicIp = publicIp
			poa.deviceInfo.PrivateIp = privateIP

			request := Request{Type: "register"}
			request.Register.DeviceInfo = poa.deviceInfo

			doc, err := json.MarshalIndent(request, "", "    ")
			if err == nil {
				token := poa.mqttClient.Publish("mine/server/request", poa.mqttQos, false, string(doc))
				token.Wait()
			} else {
				log.Println(err)
			}
		}

		// loop start
		go func() {
			ticker := time.NewTicker(time.Second * time.Duration(poa.intervalSec))
			go func() {
				for range ticker.C {
					poa.condCh <- 0
				}
			}()

			// run once at startup
			go func() {
				poa.condCh <- 0
			}()

			for {
				<-poa.condCh

				publicIp, _ := nettool.GetPublicIP()
				privateIP := nettool.GetPrivateIP()

				if poa.deviceInfo.OwnNumber <= 0 {
					poa.deviceInfo.PublicIp = publicIp
					poa.deviceInfo.PrivateIp = privateIP

					request := Request{Type: "register"}
					request.Register.DeviceInfo = poa.deviceInfo

					doc, err := json.MarshalIndent(request, "", "    ")
					if err == nil {
						token := poa.mqttClient.Publish("mine/server/request", poa.mqttQos, false, string(doc))
						token.Wait()
					} else {
						log.Println(err)
					}
				} else if publicIp != "" && privateIP != "" {
					poa.deviceInfo.PublicIp = publicIp
					poa.deviceInfo.PrivateIp = privateIP

					doc, err := json.MarshalIndent(poa.deviceInfo, "", "    ")
					if err == nil {
						token := poa.mqttClient.Publish(fmt.Sprintf("mine/%s/%s/poa/info", publicIp, poa.deviceInfo.DeviceId), poa.mqttQos, false, string(doc))
						token.Wait()

						poa.MqttPublishTimestamp = time.Now().Unix()
					} else {
						log.Println(err)
					}
				}
			}
		}()
	}()
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

		fmt.Println("set OwnNumber =", poa.deviceInfo.OwnNumber)

		poa.ForcePublish()
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
