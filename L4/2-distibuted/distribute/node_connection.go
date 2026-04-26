package distribute

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
	"time"
	"widlberries-go-course/L4-2/cut"
)

type TCPMessageType uint16

const (
	TCPMessageType_HandShake TCPMessageType = iota
	TCPMessageType_DenyConnectAsOverflow
	TCPMessageType_WorkInstruction
	TCPMessageType_WorkDone
	TCPMessageType_WorkSuggestion
	TCPMessageType_WorkSuggestionAccept
	TCPMessageType_WorkSuggestionDeny
	TCPMessageType_Disconnect
)

type TCPMessage struct {
	Type            TCPMessageType
	FollowerID      uint64
	WorkSize        uint64 // size in bytes
	FlagsRangeCount uint16
	FlagsDelimSize  uint16
	FlagsSeparated  bool
}

type Connection struct {
	ID                uint64
	TcpConnection     net.Conn
	WorkChannel       chan WorkMessage // This one only to use to send work from leader to follower
	TcpPort           uint16
	LastHeartBeatTime time.Time
}

func connectToLeader(ip uint32, tcpPort uint16) (net.Conn, error) {
	addr := fmt.Sprintf("%d.%d.%d.%d:%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip),
		tcpPort,
	)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("Failed to tpc connect to leader at %s", addr)
		return nil, err
	}
	tcpHandShakeMessage := TCPMessage{
		Type:       TCPMessageType_HandShake,
		FollowerID: nodeState.ID,
	}
	binary.Write(conn, binary.BigEndian, tcpHandShakeMessage)
	var tcpMessage TCPMessage
	log.Printf("Going to read now handshake as follower")
	err = binary.Read(conn, binary.BigEndian, &tcpMessage)
	if err != nil {
		return nil, err
	}
	if tcpMessage.Type != TCPMessageType_HandShake {
		conn.Close()
		return nil, errors.New("Handshake failed")
	}
	log.Printf("Connected to leader using tcp at %s", addr)
	return conn, nil
}

func ReadMessageFromBroadcast(listenConn *net.UDPConn) Message {
	buf := make([]byte, binary.Size(Message{}))

	listenConn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, _, err := listenConn.ReadFromUDP(buf)
	if err != nil {
		// Print if its not a timout error
		if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
			log.Printf("error from reading %v", err)
		}
		return Message{}
	}
	message := convertBytesToMessage(buf[:n])
	return message
}

func CreateListeningUDPBroadcastConnection(port int) (*net.UDPConn, error) {
	lc := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			return c.Control(func(fd uintptr) {
				syscall.SetsockoptInt(syscall.Handle(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
			})
		},
	}
	conn, err := lc.ListenPacket(context.Background(), "udp4", ":9999")
	if err != nil {
		return nil, err
	}

	return conn.(*net.UDPConn), nil
}

func startTCPListenerAsLeader(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleFollower(conn)
	}
}

func performLeaderHearthBeat(writeConn *net.UDPConn) {
	if nodeState.Type == NodeType_Leader && time.Since(nodeState.LastHearthBeat) >= timeoutHeartBeatSending {
		hearthBeatMessage := MakeMessageFilledWithNodeState()
		hearthBeatMessage.Type = Message_HearthBeat
		sendMulticastMessage(hearthBeatMessage, writeConn)
		nodeState.LastHearthBeat = time.Now()
	}
}

func sendWorkTCPMessage(workMessage WorkMessage, connection net.Conn, tcpMessageType TCPMessageType) error {
	tcpMessage := TCPMessage{
		Type:            tcpMessageType,
		WorkSize:        uint64(len(workMessage.Text)),
		FlagsRangeCount: uint16(len(workMessage.Flags.FieldRanges)),
		FlagsDelimSize:  uint16(len(workMessage.Flags.Delimeter)),
		FlagsSeparated:  workMessage.Flags.Separated,
	}
	err := binary.Write(connection, binary.BigEndian, tcpMessage)
	if err != nil {
		return err
	}

	for _, fieldRange := range workMessage.Flags.FieldRanges {
		binary.Write(connection, binary.BigEndian, int64(fieldRange.Start))
		binary.Write(connection, binary.BigEndian, int64(fieldRange.End))
	}
	connection.Write([]byte(workMessage.Flags.Delimeter))
	_, err = connection.Write([]byte(workMessage.Text))
	if err != nil {
		return err
	}

	return nil
}

// it should already read the TCPMessage before calling this one
func readWorkFromTCP(tcpMessage TCPMessage, connection net.Conn) (WorkMessage, error) {
	var joinedErrors error
	ranges := make([]cut.Range, tcpMessage.FlagsRangeCount)
	for i := range ranges {
		var start, end int64
		errors.Join(binary.Read(connection, binary.BigEndian, &start), joinedErrors)
		errors.Join(binary.Read(connection, binary.BigEndian, &end), joinedErrors)
		ranges[i] = cut.Range{Start: int(start), End: int(end)}
	}

	delimBuf := make([]byte, tcpMessage.FlagsDelimSize)
	_, err := io.ReadFull(connection, delimBuf)
	errors.Join(err, joinedErrors)
	// read work text
	workBuf := make([]byte, tcpMessage.WorkSize)
	_, err = io.ReadFull(connection, workBuf)
	errors.Join(err, joinedErrors)

	flags := cut.Flags{
		FieldRanges: ranges,
		Delimeter:   string(delimBuf),
		Separated:   tcpMessage.FlagsSeparated,
	}

	workMessage := WorkMessage{
		Type:  WorkMessageType_Start,
		Text:  string(workBuf),
		Flags: flags}

	return workMessage, joinedErrors
}

func convertBytesToMessage(data []byte) Message {
	var message Message
	buf := bytes.NewReader(data)
	err := binary.Read(buf, binary.BigEndian, &message)
	if err != nil {
		log.Printf("error in converting bytes to message: %v", err)
	}
	return message
}

func sendMulticastMessage(message Message, conn *net.UDPConn) {
	messageBytes := new(bytes.Buffer)
	err := binary.Write(messageBytes, binary.BigEndian, message)
	if err != nil {

		log.Printf("error when converting to binary broadcast message: %v", err)
		return
	}

	_, err = conn.Write(messageBytes.Bytes())
	if err != nil {
		log.Printf("error when sending broadcast message: %v", err)
		return
	}

	//Heartbeats creates too much spam
	if message.Type != Message_HearthBeat {
		log.Printf("Sended message %v", message)
	}
}
