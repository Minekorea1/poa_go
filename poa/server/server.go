package server

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	poaContext "poa/context"
	"poa/log"
	"poa/poa"
)

var logger = log.NewLogger("server")

type HttpServer struct {
	poaContext *poaContext.Context
	poa        *poa.Poa
	webContext WebContext
}

type WebContext struct {
	DeviceId             string
	MacAddress           string
	PublicIp             string
	PrivateIp            string
	Owner                string
	OwnNumber            string
	DeviceType           string
	DeviceDesc           string
	Version              string
	MqttPublishTimestamp string
	MqttBrokerAddress    string
	MqttPort             string
	MqttUser             string
	MqttPassword         string
}

func NewHttpServer(poaCtx *poaContext.Context, poa *poa.Poa) *HttpServer {
	localServer := HttpServer{poaContext: poaCtx, poa: poa}
	return &localServer
}

func (localServer *HttpServer) GetWebContext() *WebContext {
	return &localServer.webContext
}

func (localServer *HttpServer) NewHttpHandler() http.Handler {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// w.Header().Add("Content Type", "text/html")
		templates := template.Must(template.ParseFiles("www/index.html"))
		templates.Execute(w, &localServer.webContext)
	})

	router.HandleFunc("/status", func(w http.ResponseWriter, req *http.Request) {
		// w.Header().Add("Content Type", "text/html")
		templates := template.Must(template.ParseFiles("www/status.html"))
		templates.Execute(w, &localServer.webContext)
	})

	router.HandleFunc("/config", func(w http.ResponseWriter, req *http.Request) {
		// w.Header().Add("Content Type", "text/html")
		templates := template.Must(template.ParseFiles("www/config.html"))
		templates.Execute(w, &localServer.webContext)
	})

	router.HandleFunc("/apply", func(w http.ResponseWriter, req *http.Request) {
		logger.LogF("===POST===")

		body, _ := ioutil.ReadAll(req.Body)
		req.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		logger.LogW(string(body))

		webContext := new(WebContext)
		err := json.NewDecoder(req.Body).Decode(webContext)
		if err == nil {
			logger.LogD("Owner =", webContext.Owner)
			logger.LogD("OwnNumber =", webContext.OwnNumber)
			logger.LogD("DeviceDesc =", webContext.DeviceDesc)
			logger.LogD("MqttBrokerAddress =", webContext.MqttBrokerAddress)
			logger.LogD("MqttPort =", webContext.MqttPort)
			logger.LogD("MqttUser =", webContext.MqttUser)
			logger.LogD("MqttPassword =", webContext.MqttPassword)

			localServer.apply(webContext)
		} else {
			logger.LogE(err)
		}

		// templates := template.Must(template.ParseFiles("www/config.html"))
		// templates.Execute(w, &localServer.webContext)
	}).Methods("POST")

	// http.Handle("/www/", http.StripPrefix("/www/", http.FileServer(http.Dir("www"))))
	router.PathPrefix("/www/").Handler(http.StripPrefix("/www/", http.FileServer(http.Dir("www"))))

	return router
}

func (localServer *HttpServer) apply(webContext *WebContext) {
	// Owner =배신규
	// OwnNumber =1
	// DeviceDesc =개발용 PC
	// MqttBrokerAddress =minekorea.asuscomm.com
	// MqttPort =1884
	// MqttUser =mine
	// MqttPassword =minekorea@7321

	deviceInfo := localServer.poa.GetDeviceInfo()
	// oldOwner := deviceInfo.Owner
	// oldOwnNumber := deviceInfo.OwnNumber
	deviceInfo.Owner = webContext.Owner
	deviceInfo.OwnNumber, _ = strconv.Atoi(webContext.OwnNumber)
	deviceInfo.DeviceDesc = webContext.DeviceDesc
	localServer.poa.WriteDeviceInfo(deviceInfo)

	oldConfigs := localServer.poaContext.Configs
	localServer.poaContext.Configs.MqttBrokerAddress = webContext.MqttBrokerAddress
	localServer.poaContext.Configs.MqttPort, _ = strconv.Atoi(webContext.MqttPort)
	localServer.poaContext.Configs.MqttUser = webContext.MqttUser
	localServer.poaContext.Configs.MqttPassword = webContext.MqttPassword
	localServer.poaContext.WriteConfig()

	if oldConfigs != localServer.poaContext.Configs {
		// dialog.ShowInformation("접속 정보 변경", "수정 사항을 적용하려면 프로그램을 재시작 해주세요.", *window)
		logger.LogW("수정 사항을 적용하려면 프로그램을 재시작 해주세요.")
	}

	go func() {
		// if oldOwner != deviceInfo.Owner && oldOwnNumber == deviceInfo.OwnNumber {
		// 	deviceInfo.OwnNumber = 0
		// }
		localServer.poa.ForcePublish()
	}()
}
