package distribute

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"widlberries-go-course/L4-2/cut"
)

type WorkMessageType uint8

const (
	WorkMessageType_None WorkMessageType = iota
	WorkMessageType_Start
	WorkMessageType_WorkBatch
	WorkMessageType_Complete
	WorkMessageType_SuggestionDeny
	WorkMessageType_SuggestionAccept
)

type WorkMessage struct {
	ID    uint64
	Type  WorkMessageType
	Text  string
	Flags cut.Flags
}

type WorkResultType uint8

const (
	WorkResultType_Started WorkResultType = iota
	WorkResultType_Done
	WorkResultType_Failed
)

type WorkResult struct {
	ID           uint64
	Type         WorkResultType
	Result       string
	FinishedTime time.Time
}

func processSelfWork(workChannel chan WorkMessage) {
	for work := range workChannel {
		result := processWork(work)

		nodeState.FollowerChannel <- FollowerNotification{
			Type:   FollowerNotificationType_WorkDone,
			ID:     nodeState.ID,
			Result: result,
		}
		log.Printf("Sended follower notification from self work leader of work done!")
	}
}

func splitWork(workMessage WorkMessage, totalWorkers uint8) []uint64 {
	textLength := uint64(len(workMessage.Text))
	targetSize := textLength / uint64(totalWorkers)
	splitIndexes := make([]uint64, 0, totalWorkers+1)
	splitIndexes = append(splitIndexes, 0)

	for i := uint8(0); i < totalWorkers; i++ {
		lastPosition := splitIndexes[len(splitIndexes)-1]
		if lastPosition >= textLength {
			break
		}
		position := lastPosition + targetSize
		for position < textLength && workMessage.Text[position] != '\n' {
			position++
		}
		if position < textLength {
			splitIndexes = append(splitIndexes, position+1)
		} else {
			splitIndexes = append(splitIndexes, textLength)
		}
	}
	return splitIndexes

}

func processWork(workMessage WorkMessage) string {
	splittedLines := strings.Split(strings.ReplaceAll(workMessage.Text, "\r\n", "\n"), "\n") //On windows \r\n is a new line, not \n
	var stringBuilder strings.Builder
	for _, line := range splittedLines {
		cut.ProcessLine(workMessage.Flags, line, &stringBuilder)
	}

	return stringBuilder.String()
}

func NewWorkMessageFromInput(input string) (WorkMessage, error) {
	var workMessage WorkMessage
	flags := cut.GetFlags(input)
	data, err := os.ReadFile(flags.InputFile)
	if err != nil {
		return workMessage, fmt.Errorf("error reading input file: %w", err)
	}
	workMessage = WorkMessage{
		Type:  WorkMessageType_Start,
		Flags: flags,
		Text:  string(data),
	}
	return workMessage, nil

}
