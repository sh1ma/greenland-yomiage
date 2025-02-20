package handler

import (
	"github.com/bwmarrin/discordgo"
)

func (h *Handler) joinvc(s *discordgo.Session, gid string, cid string) (*discordgo.VoiceConnection, error) {
	joined := hasJoined(s, gid)

	v, err := s.ChannelVoiceJoin(gid, cid, false, false)
	if err != nil {
		return v, err
	}

	if joined {
		return v, nil
	}

	go func() {
		for x := range v.OpusRecv {
			h.soundPacket <- x
		}
	}()

	return v, nil
}

func hasJoined(session *discordgo.Session, gid string) bool {
	_, stateErr := session.State.VoiceState(gid, session.State.User.ID)

	return discordgo.ErrStateNotFound != stateErr
}
