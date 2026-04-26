package distribute

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

func doFollowerWork(leaderConnection net.Conn) {
	log.Printf("Started follower work")
	for {

		var tcpMessage TCPMessage
		err := binary.Read(leaderConnection, binary.BigEndian, &tcpMessage)
		if err != nil {
			log.Printf("Failed to read tcp in  follower work")
			return
		}
		log.Printf("As follower received tcp message from leader: %v", tcpMessage)
		switch tcpMessage.Type {
		case TCPMessageType_WorkSuggestionAccept:
			log.Printf("Suggested work was accepted by leader")
		case TCPMessageType_WorkSuggestionDeny:
			log.Printf("Suggested work was denied, leader already working")
		case TCPMessageType_WorkInstruction:

			log.Printf("We received some work instruction")
			// remove if no need to check failing nodes
			{
				log.Printf("We sleep for debug and being able to stop do work and check how failing works")
				time.Sleep(25 * time.Second)
			}
			workMessage, err := readWorkFromTCP(tcpMessage, leaderConnection)
			if err != nil {

				log.Printf("Failed to reed work instruction")
				return
			}

			workResult := processWork(workMessage)
			workResultMessage := TCPMessage{
				Type:       TCPMessageType_WorkDone,
				FollowerID: nodeState.ID,
				WorkSize:   uint64(len(workResult)),
			}
			log.Printf("We did some work")
			err = binary.Write(leaderConnection, binary.BigEndian, workResultMessage)
			if err != nil {
				return
			}

			err = binary.Write(leaderConnection, binary.BigEndian, []byte(workResult))
			if err != nil {
				return
			}

		}

	}
}

func handleWorkMessagesAsFollower() {

	for {
		select {
		// for follower we only read from input and send it as suggestion to leader
		case workMessage := <-nodeState.WorkChan:
			if workMessage.Type == WorkMessageType_Start {
				log.Printf("We got input as follower")

				if len(workMessage.Flags.OutputFile) > 0 {
					log.Printf("Output flag allowed only on Leader Node!")
					continue
				}

				// Going to suggest a work

				err := sendWorkTCPMessage(workMessage, nodeState.LeaderConnection.TcpConnection, TCPMessageType_WorkSuggestion)

				if err != nil {
					log.Printf("Failed to send Text of work suggestion to leader")
					continue
				}
			}
		default:
			return
		}
	}
}
