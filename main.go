package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"git.sr.ht/~adnano/go-gemini"
	"github.com/bwmarrin/discordgo"
)

// Bot parameters
var (
	BotToken       = flag.String("token", "", "Bot access token")
	CertPath       = flag.String("cert-path", "", "Certificate path")
	PrivateKeyPath = flag.String("private-key-path", "", "Private key path")
)

var (
	s        *discordgo.Session
	tls_cert tls.Certificate
	g_client gemini.Client
)

func init() { flag.Parse() }

func init() {
	var err error
	s, err = discordgo.New("Bot " + *BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	if *CertPath != "" {
		tls_cert, err = tls.LoadX509KeyPair(*CertPath, *PrivateKeyPath)
		if err != nil {
			log.Fatalf("Invalide certificate : %v", err)
		}
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
		"local":        "local.status()",
		"local_toggle": "local.toggle()",
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"local": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			status := localStatus()
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Status du local : " + "**" + status + "**",
				},
			})
		},
	}

	textCommandHandlers = map[string]func(s *discordgo.Session, m *discordgo.MessageCreate){
		"local": func(s *discordgo.Session, m *discordgo.MessageCreate) {
			status := localStatus()
			s.ChannelMessageSendReply(m.ChannelID, "Status du local : "+"**"+status+"**", m.Reference())
		},
		"local_toggle": func(s *discordgo.Session, m *discordgo.MessageCreate) {
			resp := toggleStatus()
			s.ChannelMessageSendReply(m.ChannelID, "Le status du local est maintenant : "+"**"+resp+"**", m.Reference())
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

	status = string(body)

	return status
}

func toggleStatus() string {
	r, err := gemini.NewRequest("gemini://status.alias-asso.fr/toggle")
	if err != nil {
		return "Erreur lors de la création de la requête."
	}
	r.Certificate = &tls_cert
	ctx := context.Background()
	resp, err := g_client.Do(ctx, r)
	if err != nil {
		return "Erreur lors de la requête"
	}
	defer resp.Body.Close()
	status := resp.Status.String()
	if status != "Redirect" {
		return "Erreur lors de la requête"
	}
	return localStatus()
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
