package handler

import (
	"github.com/bwmarrin/discordgo"
	"github.com/samber/lo"
)

type command struct {
	AppCmd  *discordgo.ApplicationCommand
	Handler func(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func (h *Handler) Interaction(dg *discordgo.Session, guildID string) (func(s *discordgo.Session, i *discordgo.InteractionCreate), []string) {
	commands := make(map[string]*command)

	commands["join"] = &command{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "join",
			Description: "あなたの参加しているボイスチャンネルに参加します。",
		},
		Handler: h.Join,
	}

	commands["leave"] = &command{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "leave",
			Description: "参加しているボイスチャンネルから退出します。",
		},
		Handler: h.Leave,
	}

	commands["add-word"] = &command{
		AppCmd: &discordgo.ApplicationCommand{
			Name:        "add-word",
			Description: "単語の読み方を登録します。",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "word",
					Description: "単語",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "pronunciation",
					Description: "読み(カタカナ)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "accent_type",
					Description: "アクセント核位置",
					Required:    true,
				},
			},
		},
		Handler: h.AddWord,
	}

	createdCommands := registerCommands(dg, guildID, lo.MapToSlice(commands, func(_ string, value *command) *discordgo.ApplicationCommand {
		return value.AppCmd
	}))

	commandIDs := lo.Map(createdCommands, func(item *discordgo.ApplicationCommand, _ int) string { return item.ID })

	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		c, ok := commands[i.ApplicationCommandData().Name]
		if !ok {
			return
		}
		c.Handler(s, i)
	}, commandIDs
}

func registerCommands(dg *discordgo.Session, guildID string, commands []*discordgo.ApplicationCommand) []*discordgo.ApplicationCommand {
	createdCommands := func() []*discordgo.ApplicationCommand {
		cmds := make([]*discordgo.ApplicationCommand, 0)
		for _, cmd := range commands {
			created, err := dg.ApplicationCommandCreate(dg.State.User.ID, guildID, cmd)
			if err != nil {
			}
			cmds = append(cmds, created)
		}
		dg.ApplicationCommandCreate(dg.State.User.ID, guildID, &discordgo.ApplicationCommand{Name: "cancel", Description: "読み上げをキャンセルします。"})
		return cmds
	}()
	return createdCommands
}
