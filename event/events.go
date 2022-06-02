package event

type EventTarget int
type EventName int

const (
	MAIN EventTarget = iota
	UPDATER
)

const (
	EVENT_MAIN_RESTART EventName = iota
	EVENT_MAIN_MQTT_CHANGE_USER_PASSWORD
	EVENT_MAIN_FORCE_UPDATE
	EVENT_MAIN_CHANGE_UPDATE_ADDRESS
)
