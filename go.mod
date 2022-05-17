module poa

go 1.18

require (
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0 // indirect
	github.com/mouuff/go-rocket-update v1.5.3 // indirect
    pao/updater v0.0.0
)

replace (
	pao/updater v0.0.0 => ./updater
)