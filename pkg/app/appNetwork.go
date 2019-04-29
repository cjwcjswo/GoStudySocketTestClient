package app

import (
	"chatClient/pkg/protocol"
	"encoding/binary"
	"errors"
	"net"
)

type packetChan struct {
	data       []byte
	finishChan chan struct{}
}

const (
	MAX_RECEIVE_BUFFER_SIZE          = 1024
	MAX_PACKET_RECV_CHAN_BUFFER_SIZE = 16
)

func (client *ChatClientApp) _initNetwork() {
	protocol.Init_packet()
	client._packetRecvChan = make(chan packetChan, MAX_PACKET_RECV_CHAN_BUFFER_SIZE)
	client._isLogin = false
	client._currentRoomNum = -1
}

func (client *ChatClientApp) _connectToServer(address string) error {
	var err error

	if client._tcpConn != nil {
		return errors.New("already connected state")
	}

	if client._tcpConn, err = net.Dial("tcp", address); err != nil {
		return err
	}

	client._startNetwork()

	return nil
}

func (client *ChatClientApp) _startNetwork() {
	client._closeNetworkChan = make(chan struct{})
	go client._packetReceive_goroutine()
	go client._socketEventProcess_goroutine()
}

func (client *ChatClientApp) _closeNetwork() {
	if client._closeNetworkChan != nil {
		close(client._closeNetworkChan)
		client._closeNetworkChan = nil
	}
	if client._tcpConn != nil {
		_ = client._tcpConn.Close()
		client._tcpConn = nil
	}

	client._isLogin = false
	client._nickname = ""
	client._currentRoomNum = -1
	client._clearUserListGroup()
}

func (client *ChatClientApp) _packetReceive_goroutine() {
	client._packetReceive_goroutineImpl()
	client._closeNetwork()
}

func (client *ChatClientApp) _packetReceive_goroutineImpl() {
	headerSize := protocol.ClientHeaderSize()
	var err error

	var startRecvPos int16
	recvBytes := 0
	receiveBuffer := make([]byte, MAX_RECEIVE_BUFFER_SIZE)

	for {
		recvBytes, err = client._tcpConn.Read(receiveBuffer[startRecvPos:])
		if err != nil {
			return
		}

		if recvBytes == 0 {
			return
		}

		readAbleByte := startRecvPos + int16(recvBytes)

		var readPos int16
		for {

			if readAbleByte < headerSize {
				break
			}

			requireDataSize := _packetTotalSize(receiveBuffer[readPos:])

			if requireDataSize > readAbleByte {
				break
			}

			packet := receiveBuffer[readPos : readPos+requireDataSize]
			readPos += requireDataSize
			readAbleByte -= requireDataSize

			finishChan := make(chan struct{})
			packetChan := packetChan{
				data:       packet,
				finishChan: finishChan,
			}
			client._packetRecvChan <- packetChan
			<-finishChan
		}

		if readAbleByte > 0 {
			copy(receiveBuffer, receiveBuffer[readPos:readPos+readAbleByte])
		}

		startRecvPos = readAbleByte
	}
}

func _packetTotalSize(data []byte) int16 {
	totalSize := binary.LittleEndian.Uint16(data)
	return int16(totalSize)
}
