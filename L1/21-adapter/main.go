package main

import (
	"fmt"
	"strings"
)

type SoundSystem interface {
	playSound(filename string)
}

type AudioSystem struct{}

func (audioSystem *AudioSystem) playSound(filename string) {
	fmt.Println("Playing file: ", filename)
}

type AudioSystemWithFormat struct{}

func (audioSystem *AudioSystemWithFormat) playSound(filename string, format string) {
	fmt.Printf("Playing format %s with file: %s\n", format, filename)
}

// Адаптер позволяет использовать уже существующий функционал, не внося изменения в него
// (также может быть + если это библиотека, к которой нет доступа)
// Минусы: добавляет еще слой абстрации, что усложняет код
type AudioSystemWithFormatAdapter struct {
	audioSystem *AudioSystemWithFormat
}

func NewAudioSystemWithFormatAdapter(audioSystem *AudioSystemWithFormat) *AudioSystemWithFormatAdapter {
	return &AudioSystemWithFormatAdapter{
		audioSystem: audioSystem,
	}
}

func (audioSystemAdapter *AudioSystemWithFormatAdapter) playSound(filename string) {
	result := strings.Split(filename, ".")
	format := result[len(result)-1]
	audioSystemAdapter.audioSystem.playSound(filename, format)
}

func playFile(soundSystem SoundSystem, filename string) {
	soundSystem.playSound(filename)
}

func main() {
	var audioSystem AudioSystem
	playFile(&audioSystem, "filename.mp3")

	audioSystemWithFormatAdapter := NewAudioSystemWithFormatAdapter(&AudioSystemWithFormat{})
	playFile(audioSystemWithFormatAdapter, "filename.mp3")
}
