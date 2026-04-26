package distribute

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

var timeoutHeartBeatSending = 1 * time.Second
var timeoutHeartBeat = 20 * time.Second
var timeoutAwaits = 15 * time.Second
var timeoutElection = 1 * time.Second
var timeoutAfterLeaderWorkFinished = 30 * time.Second

var nodeState = &NodeState{}

func SetNode() *NodeState {
	nodeState = &NodeState{}
	nodeState.ID = generateID()
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal(err)
	}
	nodeState.TCPListener = listener
	port := uint16(listener.Addr().(*net.TCPAddr).Port)
	nodeState.WorkID = generateID()
	log.Printf("TCP port: %d", port)
	nodeState.TCPPort = port

	nodeState.Status = StateStatus_Hello
	nodeState.WorkChan = make(chan WorkMessage)
	return nodeState
}

func OperateBroadcastState(listenConn *net.UDPConn, writeConn *net.UDPConn, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		receivedMessage := nodeState.PreviousMessage
		nodeState.PreviousMessage = Message{}
		if receivedMessage.Type == Message_None {
			receivedMessage = ReadMessageFromBroadcast(listenConn)
		}
		if receivedMessage.FromID == nodeState.ID {
			receivedMessage = Message{}
		}
		performStateMachine(receivedMessage, listenConn, writeConn)
		// Heartbeats messages creates too much spam
		if receivedMessage.Type != Message_None && receivedMessage.Type != Message_HearthBeat {
			log.Printf("received message was %v", receivedMessage)
		}

		// we only use heartbeath from leader so other nodes can connect to it
		// we use tcp connection to know when follower disconnected
		performLeaderHearthBeat(writeConn)
	}
}

func performStateMachine(receivedMessage Message, listenConn *net.UDPConn, writeConn *net.UDPConn) bool {
	if nodeState.PreviousDebugStatus != nodeState.Status || nodeState.PreviousDebugType != nodeState.Type {
		nodeState.PreviousDebugStatus = nodeState.Status
		nodeState.PreviousDebugType = nodeState.Type
		log.Printf("Changes in node type: %v and status %v", nodeState.Type, nodeState.Status)
	}
	isMessageConsumed := false

	// Going to election mode instantly
	if receivedMessage.Term > nodeState.Term && receivedMessage.Type == Message_Election && nodeState.Status != StateStatus_Electing {
		setStateForElection(receivedMessage)
	}

	switch nodeState.Status {
	case StateStatus_Hello:
		//We send hello and go to await mode for leader heartbeat
		helloMessage := MakeMessageFilledWithNodeState()
		helloMessage.Type = Message_Hello
		sendMulticastMessage(helloMessage, writeConn)
		nodeState.Status = StateStatus_Awaits
		nodeState.TimeoutStart = time.Now()

	case StateStatus_Awaits:
		if (receivedMessage.Type == Message_HearthBeat && receivedMessage.SenderType == NodeType_Leader || receivedMessage.Type == Message_LeaderAnnounce) && nodeState.Term <= receivedMessage.Term {
			isMessageConsumed = true
			leaderConnection, err := connectToLeader(receivedMessage.IP, receivedMessage.TCPPort)
			if err == nil {
				nodeState.Term = receivedMessage.Term
				nodeState.LeaderID = receivedMessage.FromID
				nodeState.Type = NodeType_Follower
				nodeState.Status = StateStatus_Working
				nodeState.LeaderConnection.LastHeartBeatTime = time.Now()
				nodeState.LeaderConnection.TcpConnection = leaderConnection
				go doFollowerWork(leaderConnection)
			} else {
				log.Printf("failed to connect to leader and move to the follower state! error: %v", err)
			}

		} else if receivedMessage.Type == Message_Election {
			// The election still going and the term is the same as this node
			// so we must extend the timeout
			nodeState.TimeoutStart = time.Now()

		} else if time.Since(nodeState.TimeoutStart) >= timeoutAwaits {
			nodeState.LeaderID = 0
			nodeState.Type = NodeType_Candidate
			nodeState.PreviousStatus = nodeState.Status
			nodeState.Status = StateStatus_Electing
			nodeState.JustStartedElection = true
		}

	case StateStatus_Electing:
		isMessageConsumed = performElection(receivedMessage, listenConn, writeConn)
	case StateStatus_Working:
		handleWorkMessagesAsFollower()
		if receivedMessage.Type == Message_HearthBeat && receivedMessage.FromID == nodeState.LeaderID {
			nodeState.LeaderConnection.LastHeartBeatTime = time.Now()
		}
		if time.Since(nodeState.LeaderConnection.LastHeartBeatTime) >= timeoutHeartBeat {
			setStateForElection(receivedMessage)
		}
	case StateStatus_Leading:
		performLeaderHearthBeat(writeConn)
		if receivedMessage.Type == Message_HearthBeat && receivedMessage.SenderType == NodeType_Leader {
			if receivedMessage.Term > nodeState.Term {
				nodeState.Status = StateStatus_Awaits
			} else if receivedMessage.Term == nodeState.Term {
				setStateForElection(receivedMessage)
			}
		}

		handleFollowerNotifications()
		handleWorkMessagesAsLeader()
		handleWorkProgress()
	}

	return isMessageConsumed
}

func handleFollowerNotifications() {
	for {
		select {
		case followerNotification := <-nodeState.FollowerChannel:
			switch followerNotification.Type {
			case FollowerNotificationType_Add:
				addFollowerConnection(followerNotification)
			case FollowerNotificationType_Remove:
				removeFollowerConnection(followerNotification.ID)
			case FollowerNotificationType_WorkDone:
				log.Printf("WE have work done from %d", followerNotification.ID)
				index := 0
				for i := range nodeState.FollowersConnection {
					if nodeState.FollowersConnection[i].ID == followerNotification.ID {
						index = i
						break
					}
				}
				workResult := &nodeState.WorkResultArray[index]
				log.Printf("we saving work from someone with id %v", followerNotification.ID)
				workResult.ID = followerNotification.ID
				workResult.Type = WorkResultType_Done
				workResult.Result = followerNotification.Result
				workResult.FinishedTime = time.Now()
				nodeState.finishedWorkers++
			}
		default:
			return
		}
	}
}

func handleWorkProgress() {
	if !nodeState.IsWorking {
		return
	}
	if nodeState.totalWorkers == (nodeState.finishedWorkers + nodeState.failedWorkers) {
		printFinishedWork()
		nodeState.IsWorking = false
	} else if nodeState.WorkResultArray[0].Type == WorkResultType_Done && time.Since(nodeState.WorkResultArray[0].FinishedTime) > timeoutAfterLeaderWorkFinished {
		timeTookToFinishForLeader := nodeState.workStartedSendingTime.Sub(nodeState.WorkResultArray[0].FinishedTime)
		timeForWorkTimout := nodeState.workFinishedSendingTime.Add(timeTookToFinishForLeader).Add(timeoutAfterLeaderWorkFinished)
		if time.Now().After(timeForWorkTimout) {
			if nodeState.finishedWorkers > (nodeState.totalWorkers * 2 / 3) {
				printFinishedWork()
			} else {
				log.Printf("Work task failed: only %d finished out of %d", nodeState.finishedWorkers, nodeState.totalWorkers)
			}

			nodeState.IsWorking = false
		}
	}

}

const (
	colorGreen = "\033[32m"
	colorRed   = "\033[31m"
	colorReset = "\033[0m"
)

func printFinishedWork() {
	if len(nodeState.WorkOutputFilePath) > 0 {
		file, err := os.Create(nodeState.WorkOutputFilePath)
		if err != nil {
			fmt.Printf("%s=== Failed to write to output file: %v ===%s", colorRed, err, colorReset)
			return
		}
		defer file.Close()

		for _, workResult := range nodeState.WorkResultArray {
			if workResult.Type == WorkResultType_Done {
				_, err := file.WriteString(workResult.Result)
				if err != nil {
					fmt.Printf("%s=== Failed to write to output file: %v ===%s", colorRed, err, colorReset)
					return
				}
			}
		}
		fmt.Printf("%s=== Results written to %s ===%s \n", colorGreen, nodeState.WorkOutputFilePath, colorReset)
	} else {
		fmt.Printf("%s=== Work Result ===%s\n", colorGreen, colorReset)
		for _, workResult := range nodeState.WorkResultArray {
			if workResult.Type == WorkResultType_Done {
				fmt.Print(workResult.Result)
			}
		}
		fmt.Printf("%s=== Finished Work Result ===%s\n", colorGreen, colorReset)
	}
}

func setStateForElection(receivedMessage Message) {
	if nodeState.LeaderConnection.TcpConnection != nil {
		nodeState.LeaderConnection.TcpConnection.Close()
		nodeState.LeaderConnection = Connection{}
	}
	nodeState.JustStartedElection = true
	nodeState.PreviousStatus = nodeState.Status
	nodeState.Status = StateStatus_Electing
	nodeState.PreviousMessage = receivedMessage
}

func performElection(receivedMessage Message, listenConn *net.UDPConn, writeConn *net.UDPConn) bool {
	isMessageConsumed := false
	if nodeState.JustStartedElection {
		nodeState.Term += 1
		nodeState.JustStartedElection = false
		nodeState.TimeoutStart = time.Now()
		electionMessage := MakeMessageFilledWithNodeState()
		electionMessage.Type = Message_Election
		sendMulticastMessage(electionMessage, writeConn)

	}
	if receivedMessage.Type == Message_Election {
		isMessageConsumed = true
		if receivedMessage.Term > nodeState.Term ||
			(receivedMessage.SenderType == NodeType_Leader && nodeState.Term == receivedMessage.Term) ||
			(receivedMessage.SenderType == NodeType_Backup && nodeState.Type != NodeType_Leader && nodeState.Term == receivedMessage.Term) ||
			(isBetterElectionCandidate(receivedMessage.TCPPort, receivedMessage.FromID) && nodeState.Term == receivedMessage.Term) {
			// Move to awaits
			nodeState.Status = StateStatus_Awaits
		} else {
			// We received message with Election type, but we are better candidate
			// extend the election time
			nodeState.TimeoutStart = time.Now()
		}
	}

	// If node gots to the timeout it means no one is better and no one moved it to the Awaits
	if time.Since(nodeState.TimeoutStart) >= timeoutElection {
		// Anonse as winner
		winnerMessage := MadeNodeLeaderAndGenerateMessage()
		sendMulticastMessage(winnerMessage, writeConn)

	}
	return isMessageConsumed

}

func MakeMessageFilledWithNodeState() Message {
	return Message{
		Term:       nodeState.Term,
		TCPPort:    nodeState.TCPPort,
		FromID:     nodeState.ID,
		SenderType: nodeState.Type,
		WorkID:     nodeState.WorkID,
	}
}

func MoveToAwaitsMode() {
	nodeState.Status = StateStatus_Awaits
	nodeState.IsWorking = false
}

func MadeNodeLeaderAndGenerateMessage() Message {
	nodeState.LeaderID = nodeState.ID
	if !(nodeState.Type == NodeType_Leader || nodeState.Type == NodeType_Backup) {
		// we change the work ID,because this node has not have info about work done before election
		nodeState.WorkID = generateID()
		nodeState.IsWorking = false

	}
	nodeState.Status = StateStatus_Leading
	nodeState.Type = NodeType_Leader
	nodeState.FollowersConnection = [128]Connection{}
	selfWorkChannel := make(chan WorkMessage)
	nodeState.FollowerChannel = make(chan FollowerNotification)
	nodeState.FollowersConnection[0] = Connection{
		ID:          nodeState.ID,
		WorkChannel: selfWorkChannel,
	}
	go processSelfWork(selfWorkChannel)

	go startTCPListenerAsLeader(nodeState.TCPListener)

	winnerMessage := MakeMessageFilledWithNodeState()
	winnerMessage.Type = Message_LeaderAnnounce
	return winnerMessage
}

// return true if current node is better then another candidate
func isBetterElectionCandidate(anotherCandidatePort uint16, anotherCandidateID uint64) bool {
	if nodeState.TCPPort != anotherCandidatePort {
		return nodeState.TCPPort < anotherCandidatePort
	}
	return nodeState.WorkID < anotherCandidateID
}

func generateID() uint64 {
	var b [8]byte
	rand.Read(b[:])
	return binary.BigEndian.Uint64(b[:])
}

type StateStatus uint16

const (
	StateStatus_None StateStatus = iota
	StateStatus_Hello
	StateStatus_Awaits
	StateStatus_Electing
	StateStatus_Working
	StateStatus_Leading
)

type InputState struct {
	IsWorking bool
	WorkChan  chan WorkMessage
}

type NodeState struct {
	Status              StateStatus
	PreviousStatus      StateStatus // The status should be saved before moving to election, It would need for Leader or Backup to continue work after election
	Type                NodeType
	ID                  uint64
	TCPPort             uint16
	WorkID              uint64 // Instead of having logs like in RAFT, just have a workID
	Term                uint64
	TimeoutStart        time.Time
	LeaderID            uint64
	JustStartedElection bool
	PreviousMessage     Message
	LastHearthBeat      time.Time

	LeaderConnection Connection
	TCPListener      net.Listener
	// This ones only for Leader
	// TODO Backup not implemente
	BackupConnetion     Connection
	FollowersConnection [128]Connection
	FollowerChannel     chan FollowerNotification
	WorkChan            chan WorkMessage // This channel to use from the input to current node(if its leader it would send to followers, if its a follower it will suggest work to leader)
	// work state
	IsWorking               bool
	workStartedSendingTime  time.Time
	workFinishedSendingTime time.Time
	WorkOutputFilePath      string
	totalWorkers            uint8
	finishedWorkers         uint8
	failedWorkers           uint8
	WorkResultArray         [128]WorkResult

	//For debug
	PreviousDebugStatus StateStatus // this only used to print the status when the change occure
	PreviousDebugType   NodeType
}

type NodeType uint8

const (
	NodeType_None NodeType = iota
	NodeType_Leader
	NodeType_Backup
	NodeType_Candidate
	NodeType_Follower
)

type MessageType uint8

const (
	Message_None MessageType = iota
	Message_Hello
	Message_HearthBeat
	Message_Election
	Message_LeaderAnnounce
	Message_Work
	Message_Exit
)

type FollowerNotificationType uint16

const (
	FollowerNotificationType_None FollowerNotificationType = iota
	FollowerNotificationType_Add
	FollowerNotificationType_Remove
	FollowerNotificationType_WorkOfferToLeader
	FollowerNotificationType_WorkDone
)

type FollowerNotification struct {
	Type               FollowerNotificationType
	ID                 uint64
	Result             string
	FollowerConnection Connection
}

type Message struct {
	Type          MessageType
	Term          uint64
	IP            uint32
	TCPPort       uint16
	FromID        uint64
	ToID          uint64
	SenderType    NodeType
	WorkID        uint64
	CurrentLeader uint64
}

func (t MessageType) String() string {
	switch t {
	case Message_None:
		return "None"
	case Message_Hello:
		return "Hello"
	case Message_HearthBeat:
		return "Heartbeat"
	case Message_Election:
		return "Election"
	case Message_LeaderAnnounce:
		return "LeaderAnnounce"
	case Message_Work:
		return "Work"
	case Message_Exit:
		return "Exit"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

func (t NodeType) String() string {
	switch t {
	case NodeType_None:
		return "None"
	case NodeType_Leader:
		return "Leader"
	case NodeType_Backup:
		return "Backup"
	case NodeType_Candidate:
		return "Candidate"
	case NodeType_Follower:
		return "Follower"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

func (s StateStatus) String() string {
	switch s {
	case StateStatus_None:
		return "None"
	case StateStatus_Hello:
		return "Hello"
	case StateStatus_Awaits:
		return "Awaits"
	case StateStatus_Electing:
		return "Electing"
	case StateStatus_Working:
		return "Working"
	case StateStatus_Leading:
		return "Leading"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

func (m Message) String() string {
	return fmt.Sprintf(
		"[%s term=%d from=%d to=%d port=%d sender=%s leader=%d workID=%d]",
		m.Type,
		m.Term,
		m.FromID,
		m.ToID,
		m.TCPPort,
		m.SenderType,
		m.CurrentLeader,
		m.WorkID,
	)
}
