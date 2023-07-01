package speaker

import (
	"log"
	"time"

	"github.com/chun37/greenland-yomiage/internal/usecase/tts"
	"golang.org/x/xerrors"
)

type Speaker struct {
	usecase  *tts.Usecase
	messages chan SpeechMessage
	quiet    <-chan struct{}
}

func NewSpeaker(usecase *tts.Usecase, messages chan SpeechMessage, quiet <-chan struct{}) *Speaker {
	return &Speaker{
		usecase:  usecase,
		messages: messages,
		quiet:    quiet,
	}
}

func (s *Speaker) Run() {
	isSpeaking := true
	messages := make([]SpeechMessage, 0)

	go func() {
		for {
			for isSpeaking || len(messages) == 0 {
				time.Sleep(time.Microsecond)
			}
			if err := s.do(messages[0]); err != nil {
				log.Println("failed to speak message: %+v", err)
			}
			messages = messages[1:]
		}
	}()

	for {
		select {
		case <-s.quiet:
			isSpeaking = false
		case message := <-s.messages:
			messages = append(messages, message)
		default:
			isSpeaking = true
		}
	}
}

func (s *Speaker) do(message SpeechMessage) error {
	if err := message.VoiceConnection.Speaking(true); err != nil {
		return xerrors.Errorf("Couldn't set speaking: %w", err)
	}

	done := make(chan struct{})
	opusChunks := make(chan []byte, 3)
	defer close(done)
	defer close(opusChunks)
	if err := s.usecase.Do(tts.UsecaseParam{
		Text:       message.Text,
		OpusChunks: opusChunks,
		Done:       done,
	}); err != nil {
		return xerrors.Errorf("failed to exec usecase: %w", err)
	}

	// Send not "speaking" packet over the websocket when we finish
	defer func() {
		err := message.VoiceConnection.Speaking(false)
		if err != nil {
			log.Println("Couldn't stop speaking", err)
		}
	}()

	for {
		select {
		case opus := <-opusChunks:
			message.VoiceConnection.OpusSend <- opus
		case <-done:
			return nil
		}
	}
}
