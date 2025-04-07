package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

var s *discordgo.Session

// Initialize Discord client
func init() {
	// wtf go
	var err error
	s, err = discordgo.New("Bot " + config.BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	// Slash commands definition
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "local",
			Description: "Interactions avec le statut du local",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "toggle",
					Description: "Changer le statut du local",
					Required:    false,
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "status",
					Description: "Obtenir le statut du local",
					Required:    false,
					Type:        discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	// Slash commands handlers
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){

		// local command
		"local": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			switch options[0].Name {
			case "status":
				status, err := GetStatusString()
				if err != nil {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Flags:   discordgo.MessageFlagsEphemeral,
							Content: "Une erreur s'est produite",
						},
					})
				} else {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: "Statut du local : " + "**" + status + "**",
						},
					})
				}
			case "toggle":
				hasRole := false
				for _, v := range i.Member.Roles {
					for _, role := range config.ToggleRoles {
						if v == role {
							hasRole = true
							break
						}
					}
				}
				if hasRole {
					status, err := ToggleStatus()
					if err != nil {
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{
								Flags:   discordgo.MessageFlagsEphemeral,
								Content: "Une erreur s'est produite",
							},
						})

					} else {
						s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
							Type: discordgo.InteractionResponseChannelMessageWithSource,
							Data: &discordgo.InteractionResponseData{
								Content: "Le statut du local est maintenant : " + "**" + status + "**",
							},
						})
					}
				} else {
					s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Flags:   discordgo.MessageFlagsEphemeral,
							Content: "Vous n'avez pas le rôle nécessaire pour cette commande",
						},
					})
				}
			}

		},
	}
)

// Add handlers
func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}
