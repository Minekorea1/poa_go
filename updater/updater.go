package updater

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/mouuff/go-rocket-update/pkg/provider"
	rokUpdater "github.com/mouuff/go-rocket-update/pkg/updater"
)

type VersionRO int

const (
	lt VersionRO = 1 + iota
	eq
	gt
)

type Updater struct {
	version string
	github  string
}

func NewUpdater() *Updater {
	updater := Updater{}
	return &updater
}

func (updater *Updater) Init(version, github string) {
	updater.version = version
	updater.github = github
}

func versionCompare(ver1, ver2 string) VersionRO {
	var major1, minor1, patch1 int
	var major2, minor2, patch2 int

	ver1 = ver1[1:]
	ver2 = ver2[1:]

	splitVer1 := strings.Split(ver1, ".")
	splitVer2 := strings.Split(ver2, ".")

	if len(splitVer1) == 3 && len(splitVer2) == 3 {
		major1, _ = strconv.Atoi(splitVer1[0])
		minor1, _ = strconv.Atoi(splitVer1[1])
		patch1, _ = strconv.Atoi(splitVer1[2])

		major2, _ = strconv.Atoi(splitVer2[0])
		minor2, _ = strconv.Atoi(splitVer2[1])
		patch2, _ = strconv.Atoi(splitVer2[2])

		if major1 > major2 {
			return gt
		} else if major1 < major2 {
			return lt
		} else {
			if minor1 > minor2 {
				return gt
			} else if minor1 < minor2 {
				return lt
			} else {
				if patch1 > patch2 {
					return gt
				} else if patch1 < patch2 {
					return lt
				} else {
					return eq
				}
			}
		}
	}

	return eq
}

func verify(u *rokUpdater.Updater) error {
	latestVersion, err := u.GetLatestVersion()
	if err != nil {
		return err
	}
	executable, err := u.GetExecutable()
	if err != nil {
		return err
	}
	cmd := exec.Cmd{
		Path: executable,
		Args: []string{executable, "-version"},
	}
	// Should be replaced with Output() as soon as test project is updated
	output, err := cmd.Output()
	if err != nil {
		return err
	}
	strOutput := string(output)

	if !strings.Contains(strOutput, latestVersion) {
		return errors.New("version not found in program output")
	}

	return nil
}

func (updater Updater) Update() {
	u := &rokUpdater.Updater{
		Provider: &provider.Github{
			RepositoryURL: updater.github,
			ArchiveName:   "poa.zip",
		},
		ExecutableName: fmt.Sprintf("poa_%s_%s", runtime.GOOS, runtime.GOARCH),
		Version:        updater.version,
	}

	lastestVersion, err := u.GetLatestVersion()

	fmt.Println("current Version:", u.Version, ", server Version:", lastestVersion)

	if err == nil {
		if versionCompare(u.Version, lastestVersion) == lt {
			fmt.Println("software update start")
			updateStatus, err := u.Update()
			if err != nil {
				log.Println(err)
			}
			if updateStatus == rokUpdater.Updated {
				if err := verify(u); err != nil {
					log.Println(err)
					log.Println("Rolling back...")
					u.Rollback()
					return
				}
				fmt.Println("software update complete.")
			} else {
				fmt.Println("software update failed.")
			}
		}
	} else {
		log.Println(err)
	}
}