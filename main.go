package main

import (
	"bufio"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var BotToken = flag.String("token", "", "Bot access token")

var s *discordgo.Session

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
}

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "local",
			Description: "Obtenir le status du local",
		},
	}

	textCommands = map[string]string{
		"local": "local.status()",
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"local": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			status := localStatus()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: status,
				},
			})
		},
	}

	textCommandHandlers = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate){
		"local": func(s *discordgo.Session, m *discordgo.MessageCreate) {
			status := localStatus()
			s.ChannelMessageSendReply(m.ChannelID, status, m.Reference())
		},
	}
)

func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		for i, v := range textCommands {
			if m.Content == v {
				if command, ok := textCommandHandlers[i]; ok {
					command(s, m)
				}
			}
		}
	})
}

func localStatus() string {
	resp, err := http.Get("https://status.alias-asso.fr/")

	var status string

	if err != nil {
		status = "Erreur lors de l'obtention du status du local"
	}

	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)

	body, _, err := reader.ReadLine()

	if err != nil {
		status = "Erreur lors de l'obtention du status du local"
	}

	status = "Status du local : " + "**" + string(body) + "**"

	return status
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	log.Println("Gracefully shutting down.")
}
