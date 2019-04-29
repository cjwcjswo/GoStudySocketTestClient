package app

import (
	"fyne.io/fyne"
	"fyne.io/fyne/widget"
	"net"
)

const (
	TITLE = "Chat Client"
)

type ChatClientApp struct {
	viewElement
	networkElement
}

type viewElement struct {
	_app    fyne.App
	_window fyne.Window

	_ipAddressEntry   *widget.Entry
	_portEntry        *widget.Entry
	_connectButton    *widget.Button
	_disconnectButton *widget.Button

	_nicknameEntry *widget.Entry
	_passwordEntry *widget.Entry
	_loginButton   *widget.Button

	_userListGroup     *widget.Group
	_userListContainer *fyne.Container
	_userIdList        []uint64

	_roomNumEntry    *widget.Entry
	_roomEnterButton *widget.Button
	_roomLeaveButton *widget.Button

	_chatContainer *fyne.Container

	_inputEntry *widget.Entry
	_sendButton *widget.Button
}

type networkElement struct {
	_tcpConn          net.Conn
	_packetRecvChan   chan packetChan
	_closeNetworkChan chan struct{}

	_isLogin        bool
	_nickname       string
	_currentRoomNum int32
}

func NewChatClientApp() *ChatClientApp {
	chatApp := new(ChatClientApp)
	chatApp._initView()
	chatApp._initNetwork()

	return chatApp
}

func (client *ChatClientApp) Start() {
	client._window.ShowAndRun()
}
