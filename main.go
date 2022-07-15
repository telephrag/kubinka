package main

import (
	"context"
	"discordgo"
	"kubinka/changestream"
	"kubinka/config"
	"kubinka/orchestrator"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func getLogFile(fileName string) *os.File {
	// setting up logging, for some reason it loggin wont work properly
	// if it was setup inside init()
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal("Failed to open file for logging.\n\n\n")
	}
	return f
}

func newDiscordSession(token string) (*discordgo.Session, *orchestrator.CmdOrchestrator) {
	discord, err := discordgo.New("Bot " + token) // TODO: Make to arg
	if err != nil {
		log.Fatal("Could not create session.\n\n\n")
	}
	discord.SyncEvents = false
	co := orchestrator.New()
	discord.AddHandler(co.Orchestrate) // see "notes 02" in NOTES.md

	err = discord.Open()
	if err != nil {
		log.Fatal("Could not open connection.\n\n\n")
	}

	return discord, co
}

func deleteCommands(ds *discordgo.Session) { // make stuff passed in as params
	for _, cmd := range orchestrator.Commands {
		err := ds.ApplicationCommandDelete(
			ds.State.User.ID,
			config.GuildID,
			cmd.ID,
		)
		if err != nil {
			log.Fatalf("Could not delete %q command: %v\n\n\n", cmd.Name, err)
		}
	}
}

func createCommands(ds *discordgo.Session, appId string, guildId string) {
	var err error
	for i, cmd := range orchestrator.Commands {
		orchestrator.Commands[i], err = ds.ApplicationCommandCreate(
			appId,
			guildId,
			cmd,
		)
		if err != nil {
			if i > 0 {
				deleteCommands(ds)
			}
			log.Fatalf("Failed to create command %s:\n %s\n\n\n", cmd.Name, err)
		}
	}
}

func main() {
	ds, co := newDiscordSession(config.Token)
	defer co.HaltUntilAllDone()
	defer ds.Close()
	defer deleteCommands(ds) // Removing commands on bot shutdown
	createCommands(ds, config.AppID, config.GuildID)

	log.SetOutput(getLogFile(config.LOG_FILE_NAME))
	log.Print("<<<<< SESSION STARTUP >>>>>\n")
	defer log.Print("<<<<< SESSION SHUTDOWN >>>>>\n\n\n")

	ctx, cancel := context.WithCancel(context.Background())
	go changestream.WatchEvents(ds, ctx, cancel) // Asynchronously watch events

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)

	// handling invalidation of collection at shutdown
	for {
		select {
		case <-interrupt:
			log.Println("Execution stopped by user")
			return
		case <-ctx.Done():
			log.Println("handled changestream cancelation at shutdown")
			return
		default:
		}
		time.Sleep(time.Millisecond * 500)
	}
}
