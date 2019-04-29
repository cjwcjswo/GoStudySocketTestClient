package app

import (
	"chatClient/pkg/protocol"
	"errors"
	"fmt"
	"fyne.io/fyne/dialog"
	"golang_socketGameServer_codelab/gohipernetFake"
	"strconv"
	"strings"
)

// 접속 버튼 클릭 이벤트
func (client *ChatClientApp) _onClickConnectButton() {
	if client._onClickConnectButtonImpl() {
		client._addLabelToChatBox("Connect Success.")
	}
}

func (client *ChatClientApp) _onClickConnectButtonImpl() bool {
	var err error

	if err = _checkOnClickValue(client._ipAddressEntry.Text, client._portEntry.Text); err != nil {
		dialog.ShowError(err, client._window)
		return false
	}

	if err = client._connectToServer(fmt.Sprintf("%s:%s", client._ipAddressEntry.Text, client._portEntry.Text)); err != nil {
		dialog.ShowError(err, client._window)
		return false
	}
	return true
}

// 접속 끊기 버튼 클릭
func (client *ChatClientApp) _onClickDisconnectButton() {
	if client._tcpConn != nil {
		client._closeNetwork()
	}
}

// 로그인 버튼 클릭
func (client *ChatClientApp) _onClickLoginButton() {
	if err := client._onClickLoginButtonImpl(); err != nil {
		dialog.ShowError(err, client._window)
		return
	}

	nickname := strings.TrimSpace(client._nicknameEntry.Text)
	password := strings.TrimSpace(client._passwordEntry.Text)
	var nicknameBuf [protocol.MAX_USER_ID_BYTE_LENGTH]byte
	var passwordBuf [protocol.MAX_USER_PW_BYTE_LENGTH]byte
	copy(nicknameBuf[:], nickname)
	copy(passwordBuf[:], password)

	loginReq := protocol.LoginReqPacket{
		UserID: nicknameBuf[:],
		PassWD: passwordBuf[:],
	}
	sendBuf, _ := loginReq.EncodingPacket()
	_, _ = client._tcpConn.Write(sendBuf)
	client._nickname = nickname
}

func (client *ChatClientApp) _onClickLoginButtonImpl() error {
	if err := _checkOnLoginValue(client._nicknameEntry.Text, client._passwordEntry.Text); err != nil {
		return err
	}

	if client._tcpConn == nil {
		return errors.New("please connect server first")
	}

	if client._isLogin {
		return errors.New("already login state")
	}

	return nil
}

// 방 입장 버튼 클릭
func (client *ChatClientApp) _onClickRoomEnterButton() {
	if err := client._onClickRoomEnterButtonImpl(); err != nil {
		dialog.ShowError(err, client._window)
		return
	}
}

func (client *ChatClientApp) _onClickRoomEnterButtonImpl() error {
	if err := _checkOnRoomEnterValue(client._roomNumEntry.Text); err != nil {
		return err
	}

	if client._tcpConn == nil {
		return errors.New("please connect server first")
	}

	if !client._isLogin {
		return errors.New("please login first")
	}

	if client._currentRoomNum != -1 {
		return errors.New(fmt.Sprintf("[RoomNum: %d] Already room enter", client._currentRoomNum))
	}

	roomNum, _ := strconv.Atoi(client._roomNumEntry.Text)
	enterReq := protocol.RoomEnterReqPacket{
		RoomNumber: int32(roomNum),
	}
	sendBuf, _ := enterReq.EncodingPacket()
	_, _ = client._tcpConn.Write(sendBuf)

	return nil
}

// 방 퇴장 버튼 클릭
func (client *ChatClientApp) _onClickRoomLeaveButton() {
	if err := client._onClickRoomLeaveButtonImpl(); err != nil {
		dialog.ShowError(err, client._window)
		return
	}
}

func (client *ChatClientApp) _onClickRoomLeaveButtonImpl() error {
	if client._tcpConn == nil {
		return errors.New("please connect server first")
	}

	if !client._isLogin {
		return errors.New("please login first")
	}

	if client._currentRoomNum == -1 {
		return errors.New("please room enter first")
	}

	leaveReq := protocol.RoomLeaveReqPacket{}
	sendBuf, _ := leaveReq.EncodingPacket()
	_, _ = client._tcpConn.Write(sendBuf)

	return nil
}

// 채팅 입력 버튼 클릭
func (client *ChatClientApp) _onClickSendButton() {
	if len(client._inputEntry.Text) == 0 {
		return
	}
	if err := client._onClickSendButtonImpl(); err != nil {
		dialog.ShowError(err, client._window)
		return
	}
	client._inputEntry.SetText("")
}

func (client *ChatClientApp) _onClickSendButtonImpl() error {
	if client._tcpConn == nil {
		return errors.New("please connect server first")
	}

	if !client._isLogin {
		return errors.New("please login first")
	}

	if client._currentRoomNum == -1 {
		return errors.New("please room enter first")
	}

	chatReq := protocol.RoomChatReqPacket{}
	chatReq.Msgs = []byte(client._inputEntry.Text)
	chatReq.MsgLength = int16(len(chatReq.Msgs))
	sendBuf, _ := chatReq.EncodingPacket()
	_, _ = client._tcpConn.Write(sendBuf)

	return nil
}

func _checkOnClickValue(ipAddress string, port string) error {
	if strings.TrimSpace(ipAddress) == "" || strings.TrimSpace(port) == "" {
		return errors.New("plz input IP address and port num")
	}

	return nil
}

func _checkOnLoginValue(nickname string, password string) error {
	if strings.TrimSpace(nickname) == "" || strings.TrimSpace(password) == "" {
		return errors.New("plz input nickname and password")
	}

	return nil
}

func _checkOnRoomEnterValue(roomNumber string) error {
	roomNumber = strings.TrimSpace(roomNumber)
	if roomNumber == "" {
		return errors.New("plz input room number")
	}
	_, err := strconv.Atoi(roomNumber)
	if err != nil {
		return err
	}

	return nil
}

func (client *ChatClientApp) _socketEventProcess_goroutine() {
loop:
	for {
		select {
		case packetChan := <-client._packetRecvChan:
			{
				client._distributePacketProcess(packetChan.data)
				packetChan.finishChan <- struct{}{}
			}
		case <-client._closeNetworkChan:
			{
				break loop
			}
		}
	}

	client._addLabelToChatBox("Disconnect Success.")
}

func (client *ChatClientApp) _distributePacketProcess(packetBuffer []byte) {
	packetId := protocol.PeekPacketID(packetBuffer)
	_, packetBody := protocol.PeekPacketBody(packetBuffer)
	switch packetId {
	// 로그인 응답
	case protocol.PACKET_ID_LOGIN_RES:
		{
			var loginResponse protocol.LoginResPacket
			loginResponse.DecodingPacket(packetBody)
			if loginResponse.Result == protocol.ERROR_CODE_NONE {
				client._addLabelToChatBox("Login Success!")
				client._isLogin = true
			} else {
				client._nickname = ""
				client._addLabelToChatBox("Login Fail!")
			}
		}
		// 방 입장 응답
	case protocol.PACKET_ID_ROOM_ENTER_RES:
		{
			var roomEnterResponse protocol.RoomEnterResPacket
			decodeResult := roomEnterResponse.Decoding(packetBody)
			if decodeResult {
				switch roomEnterResponse.Result {
				case protocol.ERROR_CODE_NONE:
					{
						client._addLabelToChatBox(fmt.Sprintf("[RoomNum: %d] Enter Success", roomEnterResponse.RoomNumber))
						client._currentRoomNum = roomEnterResponse.RoomNumber
						client._addUserObjectToListGroup(roomEnterResponse.RoomUserUniqueId, client._nickname)
						return
					}
				case protocol.ERROR_CODE_ENTER_ROOM_INVALID_USER_ID:
					{
						client._addLabelToChatBox("Room Enter Process - Invalid User Id!")
						return
					}
				case protocol.ERROR_CODE_ENTER_ROOM_USER_FULL:
					{
						client._addLabelToChatBox("Room is Full!!")
						return
					}
				case protocol.ERROR_CODE_ENTER_ROOM_DUPLCATION_USER:
					{
						client._addLabelToChatBox("Already Enter state this room!")
						return
					}
				case protocol.ERROR_CODE_ENTER_ROOM_INVALID_SESSION_STATE:
					{
						client._addLabelToChatBox("Enter Room State Change Fail!")
						return
					}
				}
			}
			client._addLabelToChatBox("Decode Error, Room Enter Fail")
		}

		// 에러 패킷
	case protocol.PACKET_ID_ERROR_NTF:
		{
			var errorNotify protocol.ErrorNtfPacket
			decodeResult := errorNotify.Decoding(packetBody)
			if decodeResult {
				switch errorNotify.ErrorCode {
				case protocol.ERROR_CODE_ROOM_INVALIDE_NUMBER:
					{
						client._addLabelToChatBox(fmt.Sprintf("Invalid Room Number!"))
						return
					}
				}
			}
			client._addLabelToChatBox("Decode Error, Error Notify")
		}

	case protocol.PACKET_ID_ROOM_USER_LIST_NTF:
		{
			var userListNotify protocol.RoomUserListNtfPacket
			_ = userListNotify.Decoding(packetBody)
			reader := gohipernetFake.MakeReader(userListNotify.UserList, true)
			for i := int8(0); i < userListNotify.UserCount; i++ {
				uniqueId, _ := reader.ReadU64()
				nickLength, _ := reader.ReadS8()
				nickname := reader.ReadBytes(int(nickLength))
				client._addUserObjectToListGroup(uniqueId, string(nickname))
			}
			client._refreshCanvas()
		}

	case protocol.PACKET_ID_ROOM_NEW_USER_NTF:
		{
			var newUserNotify protocol.RoomNewUserNtfPacket
			newUserNotify.Decoding(packetBody)
			reader := gohipernetFake.MakeReader(newUserNotify.User, true)
			uniqueId, _ := reader.ReadU64()
			nickLength, _ := reader.ReadS8()
			nickname := reader.ReadBytes(int(nickLength))
			client._addUserObjectToListGroup(uniqueId, string(nickname))
			client._refreshCanvas()
		}

	case protocol.PACKET_ID_ROOM_LEAVE_RES:
		{
			var leaveResult protocol.RoomLeaveResPacket
			leaveResult.Decoding(packetBody)
			if leaveResult.Result == protocol.ERROR_CODE_NONE {
				client._addLabelToChatBox(fmt.Sprintf("[RoomNum: %d] Leave Success!", client._currentRoomNum))
				client._currentRoomNum = -1
				client._clearUserListGroup()
			} else {
				client._addLabelToChatBox(fmt.Sprintf("[RoomNum: %d] Leave Fail!", client._currentRoomNum))
			}
		}

	case protocol.PACKET_ID_ROOM_LEAVE_USER_NTF:
		{
			var leaveNotify protocol.RoomLeaveUserNtfPacket
			leaveNotify.Decoding(packetBody)
			client._clearUserListByUniqueId(leaveNotify.UserUniqueId)
		}

	case protocol.PACKET_ID_ROOM_CHAT_NOTIFY:
		{
			var chatNotify protocol.RoomChatNtfPacket
			chatNotify.Decoding(packetBody)
			nickname := client._getUserNicknameByUniqueId(chatNotify.RoomUserUniqueId)
			client._addLabelToChatBox(fmt.Sprintf("[%s]: %s", nickname, chatNotify.Msg))
		}

	case protocol.PACKET_ID_ROOM_CHAT_RES:
		{
			var chatResult protocol.RoomChatResPacket
			chatResult.Decoding(packetBody)
			if chatResult.Result != protocol.ERROR_CODE_NONE {
				client._addLabelToChatBox(fmt.Sprintf("[ErrorCode: %d] Chat Send Fail", chatResult.Result))
			}
		}
	}
}

func (client *ChatClientApp) _refreshCanvas() {
	client._window.Canvas().Refresh(client._window.Content())
}
