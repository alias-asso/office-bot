package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	BotToken       string   `toml:"bot-token"`
	StatusFilePath string   `toml:"status-file-path"`
	ServerPort     string   `toml:"server-port"`
	PasswordHash   string   `toml:"password-hash"`
	ToggleRoles    []string `toml:"toggle-roles"`
}

var (
	config Config
)

// Command line parameters
var (
	ConfigFile = flag.String("config", "/etc/office-bot/config.toml", "Specify custom config file location")
)

func Configure() bool {
	flag.Parse()
	configFileContent, err := os.ReadFile(*ConfigFile)
	if err != nil {
		log.Fatalf("Config file not found.")
	}

	err = toml.Unmarshal(configFileContent, &config)
	if err != nil {
		log.Fatalf("Error parsing the config file : %s", err)
	}
	return true
}

// multiple init function in the same package are called in
// lexical order, so I need to call a Configure function
// and assign it to a global variable to make it be called
// before all init functions
var ConfigSuccess = Configure()

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
		log.Printf("Creating command '%v'...", v.Name)
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	server := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: mux,
	}
	log.Println("Server started on port " + config.ServerPort + " press Ctrl+C to exit")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	<-stop

	log.Println("Removing commands...")
	for _, v := range registeredCommands {
		err := s.ApplicationCommandDelete(s.State.User.ID, "", v.ID)
		if err != nil {
			log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
		}
	}

	statusFile.Close()

	log.Println("Gracefully shutting down")
}
