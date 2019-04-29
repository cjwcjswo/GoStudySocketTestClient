package app

import (
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/widget"
	"image/color"
)

func (client *ChatClientApp) _initView() {
	client._app = app.New()
	client._window = client._app.NewWindow(TITLE)
	client._window.SetFixedSize(true)

	client._ipAddressEntry = widget.NewEntry()
	client._ipAddressEntry.SetPlaceHolder("IP Address")
	client._ipAddressEntry.SetText("127.0.0.1")
	client._portEntry = widget.NewEntry()
	client._portEntry.SetPlaceHolder("Port Num")
	client._portEntry.SetText("11021")

	client._connectButton = widget.NewButton("Connect", client._onClickConnectButton)
	client._disconnectButton = widget.NewButton("Disconnect", client._onClickDisconnectButton)

	client._nicknameEntry = widget.NewEntry()
	client._nicknameEntry.SetPlaceHolder("Nickname")
	client._passwordEntry = widget.NewEntry()
	client._passwordEntry.SetPlaceHolder("Password")

	client._loginButton = widget.NewButton("Login", client._onClickLoginButton)

	client._roomNumEntry = widget.NewEntry()
	client._roomNumEntry.SetPlaceHolder("Room Number")

	client._roomEnterButton = widget.NewButton("Enter", client._onClickRoomEnterButton)
	client._roomLeaveButton = widget.NewButton("Leave", client._onClickRoomLeaveButton)

	client._inputEntry = widget.NewEntry()
	client._inputEntry.SetPlaceHolder("Input Chat Message")
	client._sendButton = widget.NewButton("Send", client._onClickSendButton)

	vLayout := layout.NewVBoxLayout()
	vContainer := fyne.NewContainerWithLayout(vLayout)

	vContainer.AddObject(_makeSpace(250, 0))
	vContainer.AddObject(client._ipAddressEntry)
	vContainer.AddObject(client._portEntry)
	vContainer.AddObject(client._connectButton)
	vContainer.AddObject(client._disconnectButton)
	vContainer.AddObject(client._nicknameEntry)
	vContainer.AddObject(client._passwordEntry)
	vContainer.AddObject(client._loginButton)
	vContainer.AddObject(client._roomNumEntry)

	hrContainer := fyne.NewContainerWithLayout(layout.NewHBoxLayout())
	hrContainer.AddObject(client._roomEnterButton)
	hrContainer.AddObject(client._roomLeaveButton)
	vContainer.AddObject(hrContainer)

	client._chatContainer = fyne.NewContainerWithLayout(layout.NewVBoxLayout())

	client._userListContainer = fyne.NewContainerWithLayout(layout.NewVBoxLayout())
	client._userListContainer.AddObject(_makeSpace(250, 0))
	client._userListGroup = widget.NewGroup("User List")
	client._userListContainer.AddObject(client._userListGroup)
	client._userIdList = make([]uint64, 0, 10)

	hLayout := layout.NewBorderLayout(nil, nil, vContainer, client._userListContainer)
	client._chatContainer.AddObject(_makeSpace(600, 0))
	hContainer := fyne.NewContainerWithLayout(hLayout, vContainer, client._userListContainer, client._chatContainer)

	inputLayout := layout.NewBorderLayout(nil, nil, nil, client._sendButton)
	inputContainer := fyne.NewContainerWithLayout(inputLayout, client._sendButton, client._inputEntry)

	allLayout := layout.NewVBoxLayout()
	allContainer := fyne.NewContainerWithLayout(allLayout, hContainer, inputContainer)
	client._window.SetContent(allContainer)
}

func (client *ChatClientApp) _addUserObjectToListGroup(userId uint64, nickname string) {
	client._userListContainer.AddObject(widget.NewLabel(nickname))
	client._userIdList = append(client._userIdList, userId)
}

func (client *ChatClientApp) _addLabelToChatBox(text string) {
	client._chatContainer.AddObject(widget.NewLabel(text))
	if len(client._chatContainer.Objects) > 9 {
		client._chatContainer.Objects = client._chatContainer.Objects[1:]
	}
	client._refreshCanvas()
}

func (client *ChatClientApp) _getUserNicknameByUniqueId(uniqueId uint64) string {
	userIdListLen := len(client._userIdList)
	resultIndex := -1
	for i := 0; i < userIdListLen; i++ {
		if client._userIdList[i] == uniqueId {
			resultIndex = i
			break
		}
	}
	if resultIndex == -1 {
		return ""
	}

	findIndex := 0
	for i := 0; i < len(client._userListContainer.Objects); i++ {
		obj := client._userListContainer.Objects[i]
		switch result := obj.(type) {
		case *widget.Label:
			{
				if findIndex == resultIndex {
					return result.Text
				}
				findIndex++
			}
		}
	}

	return ""
}

func (client *ChatClientApp) _clearUserListGroup() {
	for i := 0; i < len(client._userListContainer.Objects); i++ {
		switch client._userListContainer.Objects[i].(type) {
		case *widget.Label:
			{
				client._userListContainer.Objects = append(client._userListContainer.Objects[:i], client._userListContainer.Objects[i+1:]...)
				i--
			}
		}
	}
	client._userIdList = make([]uint64, 0, 10)
}

func (client *ChatClientApp) _clearUserListByUniqueId(uniqueId uint64) {
	userIdListLen := len(client._userIdList)
	resultIndex := -1
	for i := 0; i < userIdListLen; i++ {
		if client._userIdList[i] == uniqueId {
			resultIndex = i
			client._userIdList = append(client._userIdList[:i], client._userIdList[i+1:]...)
			break
		}
	}
	if resultIndex == -1 {
		return
	}

	findIndex := 0
	for i := 0; i < len(client._userListContainer.Objects); i++ {
		switch client._userListContainer.Objects[i].(type) {
		case *widget.Label:
			{
				if findIndex == resultIndex {
					client._userListContainer.Objects = append(client._userListContainer.Objects[:i], client._userListContainer.Objects[i+1:]...)
					return
				}
				findIndex++
			}
		}
	}
}

func _makeSpace(width int, height int) fyne.CanvasObject {
	space := canvas.NewRectangle(&color.RGBA{})
	space.SetMinSize(fyne.NewSize(width, height))
	return space
}
