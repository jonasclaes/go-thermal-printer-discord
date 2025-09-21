package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/jonasclaes/go-thermal-printer-discord/internal/swagger/client"
	"github.com/jonasclaes/go-thermal-printer-discord/internal/swagger/client/printer"
	"github.com/jonasclaes/go-thermal-printer-discord/internal/swagger/models"
	"github.com/spf13/viper"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "todo",
			Description: "Print a todo item to the thermal printer",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "title",
					Description: "Title of the todo item",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
				},
			},
		},
	}
	commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
		"todo": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
			options := interaction.ApplicationCommandData().Options

			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			var title string

			if option, ok := optionMap["title"]; ok {
				title = option.StringValue()
			}

			var data = ""

			data = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s\n", title)))

			transport := httptransport.New(viper.GetString("printer.host"), "", []string{viper.GetString("printer.scheme")})
			apiKeyAuth := httptransport.APIKeyAuth("x-api-key", "header", viper.GetString("printer.api_key"))
			apiClient := client.New(transport, strfmt.Default)
			_, err := apiClient.Printer.PostAPIV1PrinterPrint(&printer.PostAPIV1PrinterPrintParams{
				Request: &models.PrinterPrintDto{
					Data: &data,
				},
			}, apiKeyAuth)
			if err != nil {
				log.Printf("error sending print job: %v", err)
				session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf(":cross_mark: Failed to send '%s' to the printer!", title),
					},
				})
				return
			}

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf(":white_check_mark: '%s' has been sent to the printer!", title),
				},
			})
		},
	}
)

func main() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	discord, err := discordgo.New(fmt.Sprintf("Bot %s", viper.GetString("discord.token")))
	if err != nil {
		log.Fatalf("error creating Discord session, %s", err)
		return
	}

	discord.AddHandler(func(session *discordgo.Session, ready *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", session.State.User.Username, session.State.User.Discriminator)
	})

	discord.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
			handler(session, interaction)
		}
	})

	// Open a websocket connection to Discord and begin listening.
	err = discord.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := discord.ApplicationCommandCreate(discord.State.User.ID, viper.GetString("discord.guild_id"), v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}
