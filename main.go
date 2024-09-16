package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"

	"git.sr.ht/~adnano/go-gemini"
	"github.com/bwmarrin/discordgo"
	"github.com/pelletier/go-toml/v2"
)

var ToggleRolesId = [2]string{
	"1042488433899217016", // team
	"691999377639866388",  // CA
}

type Config struct {
	BotToken       string `toml:"bot-token"`
	CertPath       string `toml:"cert-path"`
	PrivateKeyPath string `toml:"private-key-path"`
}

// Bot parameters
var (
	ConfigFile = flag.String("config", "/etc/office-bot/config.toml", "Specify custom config file location")
)

// Clients initialization
var (
	s        *discordgo.Session
	tls_cert tls.Certificate
	g_client gemini.Client
)

func init() { flag.Parse() }

func init() {
	configFileContent, err := os.ReadFile(*ConfigFile)
	if err != nil {
		log.Fatalf("Config file not found.")
	}

	var config Config
	err = toml.Unmarshal(configFileContent, &config)
	if err != nil {
		log.Fatalf("Error parsing the config file : %s", err)
	}

	s, err = discordgo.New("Bot " + config.BotToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	if config.CertPath != "" {
		tls_cert, err = tls.LoadX509KeyPair(config.CertPath, config.PrivateKeyPath)
		if err != nil {
			log.Fatalf("Invalid certificate: %v", err)
		}
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
		"local": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			switch options[0].Name {
			case "status":
				status, err := localStatus()
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
							Content: "Statut du local: " + "**" + status + "**",
						},
					})
				}
			case "toggle":
				hasRole := false
				for _, v := range i.Member.Roles {
					for _, role := range ToggleRolesId {
						if v == role {
							hasRole = true
							break
						}
					}
				}
				if hasRole {
					resp, err := toggleStatus()
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
								Content: "Le statut du local est maintenant: " + "**" + resp + "**",
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

func localStatus() (string, error) {
	resp, err := g_client.Get(context.Background(), "gemini://status.alias-asso.fr/")
	if err != nil {
		return "", errors.New("Erreur lors de l'obtention du statut du local: la requête a échoué")
	}

	defer resp.Body.Close()
	reader := bufio.NewReader(resp.Body)

	//Read lines while it starts with #
	line, _, err := reader.ReadLine()
	for line[0] == '#' {
		line, _, err = reader.ReadLine()
	}
	status := strings.Split(string(line), " ")[4]

	if err != nil {
		return "", errors.New("Erreur lors de l'obtention du statut du local: impossible de lire la réponse à la requête")
	}

	return status, nil

}

func toggleStatus() (string, error) {
	r, err := gemini.NewRequest("gemini://status.alias-asso.fr/toggle")
	if err != nil {
		return "", errors.New("Erreur lors de la création de la requête Gemini")
	}
	r.Certificate = &tls_cert
	ctx := context.Background()
	resp, err := g_client.Do(ctx, r)
	if err != nil {
		return "", errors.New("Erreur lors de la requête Gemini")
	}
	defer resp.Body.Close()
	status := resp.Status.String()
	if status != "Redirect" {
		return "", errors.New("Erreur lors de la requête Gemini: code de retour inattendue")
	}
	localStatusString, err := localStatus()
	if err != nil {
		return "", err
	}
	return localStatusString, nil
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open Discord session: %v", err)
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

	log.Println("Gracefully shutting down")
}
