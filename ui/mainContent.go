package ui

import (
	"fmt"
	"poa/poa"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type CustomContent interface {
	GetContent() *fyne.Container
}

type Menu struct {
	Title, Intro string
	content      CustomContent
}

var (
	poaInst *poa.Poa

	menus     map[string]Menu
	menuIndex map[string][]string

	Status *contentStatus
	Config *contentConfig
)

type contentStatus struct {
	Content          *fyne.Container
	labelOwner       *widget.Label
	labelOwnNumber   *widget.Label
	labelDesc        *widget.Label
	labelPublicIp    *widget.Label
	labelPrivateIp   *widget.Label
	labelMacAddress  *widget.Label
	labelDeviceId    *widget.Label
	labelLastPoaTime *widget.Label
}

type contentConfig struct {
	Content     *fyne.Container
	ownerEntry  *widget.Entry
	ownNumber   *numericalEntry
	description *widget.Entry
}

func Init(_ *fyne.App, p *poa.Poa) {
	poaInst = p

	Status = newStatus()
	Config = newConfig()

	menus = map[string]Menu{
		"status":  {"상태", "현재 상태를 표시합니다.", Status},
		"configs": {"설정", "현재 장치에 대한 설정을 할 수 있습니다.", Config},
	}

	menuIndex = map[string][]string{
		"": {"status", "configs"},
		// "collections": {"list", "table", "tree"},
	}
}

func (menu *Menu) MakeMenu(main *fyne.Container) *fyne.Container {
	tree := &widget.Tree{
		ChildUIDs: func(uid string) []string {
			return menuIndex[uid]
		},
		IsBranch: func(uid string) bool {
			children, ok := menuIndex[uid]

			return ok && len(children) > 0
		},
		CreateNode: func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Collection Widgets")
		},
		UpdateNode: func(uid string, branch bool, obj fyne.CanvasObject) {
			t, ok := menus[uid]
			if !ok {
				fyne.LogError("Missing panel: "+uid, nil)
				return
			}
			obj.(*widget.Label).SetText(t.Title)
		},
		OnSelected: func(uid string) {
			if m, ok := menus[uid]; ok {
				main.Objects = []fyne.CanvasObject{m.content.GetContent()}

				if uid == "configs" {
					deviceInfo := poaInst.GetDeviceInfo()
					Config.ownerEntry.SetText(deviceInfo.Owner)
					Config.ownNumber.SetText(strconv.FormatInt(int64(deviceInfo.OwnNumber), 10))
					Config.description.SetText(deviceInfo.DeviceDesc)
				}
			}
		},
	}

	return container.NewMax(tree)
}

func newStatus() *contentStatus {
	status := contentStatus{}

	status.Content = container.NewVBox()

	status.labelOwner = widget.NewLabel("사용자:")
	status.labelOwnNumber = widget.NewLabel("장치 번호:")
	status.labelDesc = widget.NewLabel("설명:")
	status.labelPublicIp = widget.NewLabel("공인IP:")
	status.labelPrivateIp = widget.NewLabel("내부IP:")
	status.labelMacAddress = widget.NewLabel("맥주소:")
	status.labelDeviceId = widget.NewLabel("장치 고유번호:")
	status.labelLastPoaTime = widget.NewLabel("마지막 통신 시간:")

	status.Content.Add(container.NewVBox(status.labelOwner, status.labelOwnNumber))
	status.Content.Add(status.labelDesc)
	status.Content.Add(container.NewVBox(widget.NewSeparator(), status.labelPublicIp, status.labelPrivateIp, status.labelMacAddress, status.labelDeviceId, status.labelLastPoaTime))

	return &status
}

func (status *contentStatus) Refresh() {
	deviceInfo := poaInst.GetDeviceInfo()

	status.labelOwner.SetText(fmt.Sprintf("사용자: %s", deviceInfo.Owner))
	status.labelOwnNumber.SetText(fmt.Sprintf("장치 번호: %d", deviceInfo.OwnNumber))
	status.labelDesc.SetText(fmt.Sprintf("설명:\n%s", deviceInfo.DeviceDesc))
	status.labelPublicIp.SetText(fmt.Sprintf("공인IP: %s", deviceInfo.PublicIp))
	status.labelPrivateIp.SetText(fmt.Sprintf("내부IP: %s", deviceInfo.PrivateIp))
	status.labelMacAddress.SetText(fmt.Sprintf("맥주소: %s", deviceInfo.MacAddress))
	status.labelDeviceId.SetText(fmt.Sprintf("장치 고유번호: %s", deviceInfo.DeviceId))
	status.labelLastPoaTime.SetText(fmt.Sprintf("마지막 통신 시간: %s", time.Unix(deviceInfo.Timestamp, 0).Format("2006-01-02 15:04")))
}

func (status *contentStatus) GetContent() *fyne.Container {
	return status.Content
}

func newConfig() *contentConfig {
	config := contentConfig{}

	config.Content = container.NewPadded()

	config.ownerEntry = widget.NewEntry()
	config.ownNumber = NewNumericalEntry()
	config.description = widget.NewEntry() //widget.NewMultiLineEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "사용자", Widget: config.ownerEntry},
			{Text: "장치 번호", Widget: config.ownNumber},
			{Text: "설명", Widget: config.description}},
		OnSubmit: func() {
			deviceInfo := poaInst.GetDeviceInfo()
			deviceInfo.Owner = config.ownerEntry.Text
			deviceInfo.OwnNumber, _ = strconv.Atoi(config.ownNumber.Text)
			deviceInfo.DeviceDesc = config.description.Text

			poaInst.WriteDeviceInfo(deviceInfo)

			go func() {
				poaInst.ForcePublish()
			}()
		},
		SubmitText: "저장",
	}

	config.Content.Add(form)

	return &config
}

func (config *contentConfig) GetContent() *fyne.Container {
	return config.Content
}
