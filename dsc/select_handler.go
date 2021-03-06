package dsc

import (
	"discordgo"
	"log"
)

func Select(s *discordgo.Session, i *discordgo.InteractionCreate) {
	h, ok := HandlerToCmd[i.ApplicationCommandData().Name]
	if !ok {
		log.Println("Couldn't retreive handler for command: ", i.ApplicationCommandData().Name)
		return
	}
	h(s, i)
}
