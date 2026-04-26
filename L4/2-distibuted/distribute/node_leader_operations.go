package distribute

import (
	"encoding/binary"
	"io"
	"log"
	"net"
	"time"
)

func handleFollower(followerConnection net.Conn) {
	var tcpMessage TCPMessage
	err := binary.Read(followerConnection, binary.BigEndian, &tcpMessage)
	if err != nil {
		log.Printf("handshake failed: %v", err)
		followerConnection.Close()
		return
	}
	// First message should be a handshake
	if tcpMessage.Type != TCPMessageType_HandShake {
		log.Printf("Should be Handshake message, but instead: %v", tcpMessage.Type)
		followerConnection.Close()
		return
	}

	workChannel := make(chan WorkMessage, 1)
	id := tcpMessage.FollowerID

	followerNotificationAdd := FollowerNotification{
		Type: FollowerNotificationType_Add,
		FollowerConnection: Connection{
			ID:            id,
			TcpConnection: followerConnection,
			WorkChannel:   workChannel, // probably I should remove workChannel
		},
	}
	nodeState.FollowerChannel <- followerNotificationAdd

}

func addFollowerConnection(followerNotification FollowerNotification) {
	emptyIndex := 0
	for i, followerConn := range nodeState.FollowersConnection {
		if followerConn.ID == 0 {
			emptyIndex = i
			break
		}
	}

	// on zero always self(leader)
	followerConnection := followerNotification.FollowerConnection
	if emptyIndex == 0 {
		// no empty space found
		denyMessage := TCPMessage{Type: TCPMessageType_DenyConnectAsOverflow}
		binary.Write(followerConnection.TcpConnection, binary.BigEndian, denyMessage)
		followerConnection.TcpConnection.Close()
		close(followerConnection.WorkChannel)
	} else {

		tcpMessage := TCPMessage{
			Type: TCPMessageType_HandShake,
		}

		err := binary.Write(followerConnection.TcpConnection, binary.BigEndian, tcpMessage)
		log.Printf("sended handshake from leader to follower")
		if err != nil {
			followerConnection.TcpConnection.Close()
			close(followerConnection.WorkChannel)
			return
		}

		nodeState.FollowersConnection[emptyIndex] = followerConnection

		go handleMessagesToFollowersWorkChan(followerConnection.WorkChannel, followerConnection.TcpConnection, followerConnection.ID)

		go handleFollowerMessagesFromTCP(followerConnection.TcpConnection, followerConnection.WorkChannel, followerConnection.ID)

		log.Printf("follower connected: %d", tcpMessage.FollowerID)
	}
}

func removeFollowerConnection(id uint64) {
	index := 0
	for i := range nodeState.FollowersConnection {
		if nodeState.FollowersConnection[i].ID == id {
			index = i

		}
	}
	nodeState.FollowersConnection[index].TcpConnection.Close()
	nodeState.FollowersConnection[index] = Connection{}
	if nodeState.IsWorking {
		workIndex := 0
		isFound := false
		for i, workResult := range nodeState.WorkResultArray {
			if workResult.ID == id {
				isFound = true
				workIndex = i
				break
			}
		}
		if isFound {
			nodeState.WorkResultArray[workIndex].Type = WorkResultType_Failed
			nodeState.failedWorkers++
		}

	}
	log.Printf("follower %d removed from slot %d", id, index)
}

func handleWorkMessagesAsLeader() {

	for {
		select {
		case workMessage := <-nodeState.WorkChan:
			if workMessage.Type == WorkMessageType_Start {

				log.Printf("We got input")
				if nodeState.IsWorking {
					log.Printf("Input Denied, already working")
					if workMessage.ID != 0 {
						for _, followerConnection := range nodeState.FollowersConnection {
							if workMessage.ID == followerConnection.ID {
								followerConnection.WorkChannel <- WorkMessage{
									Type: WorkMessageType_SuggestionDeny,
								}
							}
						}

					}
					return
				}
				if len(workMessage.Text) == 0 {
					return
				}
				// init the work
				nodeState.IsWorking = true
				nodeState.workStartedSendingTime = time.Now()
				nodeState.WorkOutputFilePath = workMessage.Flags.OutputFile
				if workMessage.ID != 0 {
					for _, followerConnection := range nodeState.FollowersConnection {
						if workMessage.ID == followerConnection.ID {
							followerConnection.WorkChannel <- WorkMessage{
								Type: WorkMessageType_SuggestionAccept,
							}
						}
					}

				}

				totalWorkers := uint8(0)
				// lets count how much connections we have right now  so we can split work properly
				for _, followerConnnection := range nodeState.FollowersConnection {
					if followerConnnection.ID != 0 {
						totalWorkers++
					}
				}
				nodeState.finishedWorkers = 0
				nodeState.WorkResultArray = [128]WorkResult{}
				indexesForFollowers := splitWork(workMessage, totalWorkers)
				nodeState.totalWorkers = uint8(len(indexesForFollowers) - 1)

				for i, follower := range nodeState.FollowersConnection {
					if i+1 >= len(indexesForFollowers) {
						break
					}
					message := WorkMessage{
						Type:  WorkMessageType_WorkBatch,
						Flags: workMessage.Flags,
					}
					log.Printf("Amount of indexes: %v", len(indexesForFollowers))
					message.Text = workMessage.Text[indexesForFollowers[i]:indexesForFollowers[i+1]]
					log.Printf("Putting message in work chan: %v", len(indexesForFollowers))
					nodeState.WorkResultArray[i] = WorkResult{
						ID: follower.ID,
					}
					follower.WorkChannel <- message
				}
				// Not ideal
				nodeState.workFinishedSendingTime = time.Now()
			}

		default:
			return
		}
	}
}

func handleMessagesToFollowersWorkChan(workChannel chan WorkMessage, followerConnection net.Conn, followerID uint64) {
	for work := range workChannel {
		switch work.Type {
		case WorkMessageType_SuggestionDeny:
			tcpMessage := TCPMessage{
				Type:     TCPMessageType_WorkSuggestionDeny,
				WorkSize: uint64(len(work.Text)),
			}
			err := binary.Write(followerConnection, binary.BigEndian, tcpMessage)
			if err != nil {
				return
			}
		case WorkMessageType_SuggestionAccept:
			tcpMessage := TCPMessage{
				Type:     TCPMessageType_WorkSuggestionAccept,
				WorkSize: uint64(len(work.Text)),
			}
			err := binary.Write(followerConnection, binary.BigEndian, tcpMessage)
			if err != nil {
				return
			}

		case WorkMessageType_WorkBatch:
			err := sendWorkTCPMessage(work, followerConnection, TCPMessageType_WorkInstruction)
			if err != nil {
				return
			}
		}
		log.Printf("sent work to follower %d", followerID)
	}
}

// handles the work we receive from follower on tpc leader
func handleFollowerMessagesFromTCP(followerConnection net.Conn, workChannel chan WorkMessage, followerID uint64) {
	followerNotificationRemove := FollowerNotification{
		Type: FollowerNotificationType_Remove,
		ID:   followerID,
	}
	for {
		// read header
		var tcpMessage TCPMessage
		err := binary.Read(followerConnection, binary.BigEndian, &tcpMessage)
		if err != nil || tcpMessage.Type == TCPMessageType_Disconnect {
			nodeState.FollowerChannel <- followerNotificationRemove
			close(workChannel)
			followerConnection.Close()
			return
		}
		switch tcpMessage.Type {
		case TCPMessageType_WorkDone:
			// read result string
			resultBuf := make([]byte, tcpMessage.WorkSize)
			_, err = io.ReadFull(followerConnection, resultBuf)
			if err != nil {
				nodeState.FollowerChannel <- followerNotificationRemove
				close(workChannel)
				followerConnection.Close()
				return
			}
			// TODO proper work implementation
			nodeState.FollowerChannel <- FollowerNotification{
				Type:   FollowerNotificationType_WorkDone,
				ID:     followerID,
				Result: string(resultBuf),
			}
		case TCPMessageType_WorkSuggestion:
			{
				workMessage, err := readWorkFromTCP(tcpMessage, followerConnection)
				if err != nil {
					nodeState.FollowerChannel <- followerNotificationRemove
					close(workChannel)
					followerConnection.Close()
					return
				}
				nodeState.WorkChan <- workMessage
			}
		}
	}
}
