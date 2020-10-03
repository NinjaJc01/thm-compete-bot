package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/robfig/cron"
)

var (
	prefix string
	users  []string
)

//Config defines the structure for a config file including token and list of THM usernames
type Config struct {
	Prefix    string   `json:"prefix"`
	Token     string   `json:"token"`
	UserList  []string `json:"users"`
	ChannelID string   `json:"channelID"`
}

//Statistics is the format for a user's stats, similar to how it's pulled from the API
type Statistics struct {
	Username string `json:"username"`
	Rank     int    `json:"userRank"`
	Points   int    `json:"points"`
}

func main() {
	//Load config
	log.Println("Parsing config")
	config := readConfig()
	if config.Token == "" || config.Prefix == "" || config.ChannelID == "" || config.UserList == nil {
		log.Println("Couldn't parse config, potentially missing values?")
		return
	}
	prefix = config.Prefix
	users = config.UserList

	//Set up discord
	discord, err := discordgo.New("Bot " + config.Token)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	discord.AddHandler(messageHandler)

	//Connect to discord
	log.Println("Bot loading...")
	err = discord.Open()
	if err != nil {
		log.Fatalln("Error: Couldn't open connection.", err.Error())
	}

	//Start timer so we can do stats every x
	log.Println("Starting timer for stats")
	c := cron.New()
	c.AddFunc("00 22 * * *", func() {
		dailyStats(discord, config.ChannelID)
	})
	c.Start()
	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	discord.Close()
}

func readConfig() (config Config) {
	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		log.Println(err.Error())
		return
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	data, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		log.Println(err.Error())
		return
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		log.Println(err.Error())
		return
	}
	return
}

func parseCompletedRooms(roomsJSON []byte) int {
	var rooms []map[string]string
	json.Unmarshal(roomsJSON, &rooms)
	return len(rooms)
}

func getUserTHMStats(username string) Statistics {
	userInfoRes, err := http.Get("https://tryhackme.com/api/user/" + username)
	if err != nil {
		log.Println(err.Error())
		return Statistics{}
	}
	userInfoResBody, err := ioutil.ReadAll(userInfoRes.Body)
	if err != nil {
		log.Println(err.Error())
		return Statistics{}
	}
	var userResStats Statistics
	err = json.Unmarshal(userInfoResBody, &userResStats)
	if err != nil {
		log.Println(err.Error())
		return Statistics{}
	}
	userResStats.Username = username
	return userResStats
}

func userStatsToField(stats Statistics) discordgo.MessageEmbedField {
	embedField := discordgo.MessageEmbedField{
		Name:  stats.Username,
		Value: fmt.Sprintf("Rank:\t%v\nPoints:\t%v", stats.Rank, stats.Points),
	}
	return embedField
}

func dailyStats(s *discordgo.Session, channelID string) {
	log.Println("Starting daily stats")
	var userStats []Statistics
	for _, user := range users {
		userStats = append(userStats, getUserTHMStats(user))
	}
	sort.Slice(userStats, func(i, j int) bool {
		return userStats[i].Rank < userStats[j].Rank
	})
	messageEmbed := discordgo.MessageEmbed{
		Title: "__Daily Stats__",
		Fields: func() []*discordgo.MessageEmbedField {
			var embedFields []*discordgo.MessageEmbedField
			for _, user := range userStats {
				currentEmbed := userStatsToField(user)
				embedFields = append(embedFields, &currentEmbed)
			}
			return embedFields
		}(),
	}
	s.ChannelMessageSendEmbed(channelID, &messageEmbed)
}

func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	//Ignore messages from the bot itself
	startTime := time.Now()
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Author.Bot { //Ignore other bots
		return
	}
	if !strings.HasPrefix(m.Content, prefix) {
		return
	}
	command := strings.Split(m.Content, " ")
	switch command[0] {
	case prefix + "stats":
		if len(command) == 2 {
			messageEmbed := discordgo.MessageEmbed{
				Title: "__User Stats__",
				Fields: func() []*discordgo.MessageEmbedField {
					var embedFields []*discordgo.MessageEmbedField
					currentEmbed := userStatsToField(getUserTHMStats(command[1]))
					return append(embedFields, &currentEmbed)
				}(),
			}
			s.ChannelMessageSendEmbed(m.ChannelID, &messageEmbed)
		}
	default:
		return
	}

	log.Println("Time to process:", time.Now().Sub(startTime))
}
