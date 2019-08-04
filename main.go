package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

var (
	botID                       string
	stopStatusRotation          chan bool
	stopScanMembers             chan bool
	stopUpdateOfficialUsers     chan bool
	kickChannel                 chan Kick
	impostorNotificationChannel chan *discordgo.Member
)

var vertcoinGuildID = "365207777406091266"
var protectedRoles = []string{"Development_team", "Mod_team", "Marketing_team", "Vertcoin Foundation"}
var protectedRoleIDs = []string{}
var protectedDiscriminatorRoles = []string{"Helper", "Rockstar"}
var protectedDiscriminatorRoleIDs = []string{}
var notifiedImpostors = []string{}
var notificationChannel = "moderation-team"
var notificationChannelID = ""
var qualifyingMemberRoles = []string{"Development_team", "Mod_team", "Marketing_team", "Vertcoin Foundation", "Helper", "Rockstar", "Vertopian", "Dyno", "Dyno Premium", "Vertcoin Discord Bot"}
var qualifyingMemberRoleIDs = []string{}

type Kick struct {
	Reason string
	Member *discordgo.Member
}

func main() {
	f, err := os.OpenFile("discord-bot.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(io.MultiWriter(f, os.Stdout))

	memberList = make([]*discordgo.Member, 0)
	stopStatusRotation = make(chan bool)
	stopScanMembers = make(chan bool)
	stopUpdateOfficialUsers = make(chan bool)
	impostorNotificationChannel = make(chan *discordgo.Member)
	kickChannel = make(chan Kick)
	discord, err := discordgo.New("Bot " + os.Getenv("DISCORD_BOT_KEY"))
	errCheck("error creating discord session", err)
	user, err := discord.User("@me")
	errCheck("error retrieving account", err)

	botID = user.ID
	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		roles, _ := discord.GuildRoles(vertcoinGuildID)
		for _, r := range roles {
			for _, pr := range protectedRoles {
				if r.Name == pr {
					protectedRoleIDs = append(protectedRoleIDs, r.ID)
				}
			}
			for _, pr := range protectedDiscriminatorRoles {
				if r.Name == pr {
					protectedDiscriminatorRoleIDs = append(protectedDiscriminatorRoleIDs, r.ID)
				}
			}
			for _, qr := range qualifyingMemberRoles {
				if r.Name == qr {
					qualifyingMemberRoleIDs = append(qualifyingMemberRoleIDs, r.ID)
				}
			}
		}

		chans, _ := discord.GuildChannels(vertcoinGuildID)
		for _, c := range chans {
			if c.Name == notificationChannel {
				notificationChannelID = c.ID
				log.Printf("Using channel %s (%s) for notifications", c.Name, c.ID)
			}
		}

	})
	discord.AddHandler(commandHandler)
	discord.AddHandler(scanMemberList)
	discord.AddHandler(rotateStatus)
	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		go func() {
			for i := range impostorNotificationChannel {
				nick := i.Nick
				if nick == "" {
					nick = i.User.Username
				}
				discord.ChannelMessageSend(notificationChannelID, fmt.Sprintf("Found impostor %s#%s (username %s#%s) (id %s)", i.Nick, i.User.Discriminator, i.User.Username, i.User.Discriminator, i.User.ID))
			}
		}()
	})

	discord.AddHandler(func(discord *discordgo.Session, ready *discordgo.Ready) {
		go func() {
			for k := range kickChannel {
				msg := fmt.Sprintf("Kicked user %s#%s (id %s) - %s", k.Member.User.Username, k.Member.User.Discriminator, k.Member.User.ID, k.Reason)
				discord.GuildMemberDeleteWithReason(vertcoinGuildID, k.Member.User.ID, k.Reason)
				discord.ChannelMessageSend(notificationChannelID, msg)
				log.Printf("%s\n", msg)
			}
		}()
	})

	err = discord.Open()
	errCheck("Error opening connection to Discord", err)
	defer discord.Close()

	stop := make(chan bool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Printf("Caught signal ^C, stopping sub processes...")
			select {
			case stopScanMembers <- true:
			default:
			}
			log.Printf("Stopped scan members")
			select {
			case stopStatusRotation <- true:
			default:
			}
			log.Printf("Stopped status rotation")
			select {
			case stopUpdateOfficialUsers <- true:
			default:
			}
			log.Printf("Stopped update loop for official users")
			stop <- true
		}
	}()

	<-stop

}

func rotateStatus(discord *discordgo.Session, ready *discordgo.Ready) {
	statuses := []string{"Wen moon", "Wen binance", "Scanning for scammers", "VTC: $8000 jk lol"}
	go func() {
		for {
			err := discord.UpdateStatus(0, statuses[rand.Intn(len(statuses)-1)])
			if err != nil {
				log.Println("Error attempting to set my status")
			}

			select {
			case <-stopStatusRotation:
				return
			case <-time.After(time.Minute):
			}
		}
	}()
}

type OfficialUser struct {
	Nick              string
	UserName          string
	UserID            string
	Discriminator     string
	DiscriminatorOnly bool
	NamesToCheck      []string
}

var officialUsers = []OfficialUser{}
var officialDiscriminatorUsers = []OfficialUser{}

var memberList []*discordgo.Member
var memberListLock = sync.Mutex{}

func updateProtectedUsers() {
	for {
		log.Printf("Recreating protected users list...\n")
		protectedMemberList := []*discordgo.Member{}
		protectedDiscriminatorMemberList := []*discordgo.Member{}

		for _, m := range memberList {
			for _, pr := range protectedRoleIDs {
				for _, r := range m.Roles {
					if r == pr {
						alreadyAdded := false
						for _, pm := range protectedMemberList {
							if pm.User.ID == m.User.ID {
								alreadyAdded = true
								break
							}
						}

						if !alreadyAdded {
							// This user should be protected!
							protectedMemberList = append(protectedMemberList, m)
						}
					}
				}
			}

			for _, pr := range protectedDiscriminatorRoleIDs {
				for _, r := range m.Roles {
					if r == pr {
						alreadyAdded := false
						for _, pm := range protectedDiscriminatorMemberList {
							if pm.User.ID == m.User.ID {
								alreadyAdded = true
								break
							}
						}

						if !alreadyAdded {
							// This user should be protected!
							protectedDiscriminatorMemberList = append(protectedDiscriminatorMemberList, m)
						}
					}
				}
			}
		}

		newOfficialUsers := make([]OfficialUser, 0)
		for _, m := range protectedMemberList {
			discriminator := m.User.Discriminator
			userID := m.User.ID

			namesToCheck := getAlikeNames(m.User.Username)
			if m.Nick != "" {
				namesToCheck = append(namesToCheck, getAlikeNames(m.Nick)...)
			}

			log.Printf("Adding user %s / %s (%s) and %d variations to official user list\n", m.Nick, m.User.Username, userID, len(namesToCheck))

			newOfficialUsers = append(newOfficialUsers, OfficialUser{
				Nick:              m.Nick,
				UserName:          m.User.Username,
				UserID:            userID,
				Discriminator:     discriminator,
				DiscriminatorOnly: false,
				NamesToCheck:      namesToCheck,
			})
		}

		for _, m := range protectedDiscriminatorMemberList {

			discriminator := m.User.Discriminator
			userID := m.User.ID
			namesToCheck := getAlikeNames(m.User.Username)
			if m.Nick != "" {
				namesToCheck = append(namesToCheck, getAlikeNames(m.Nick)...)
			}

			log.Printf("Adding user %s / %s (%s) and %d variations to official user list (discriminator match only)\n", m.Nick, m.User.Username, userID, len(namesToCheck))

			newOfficialUsers = append(newOfficialUsers, OfficialUser{
				Nick:              m.Nick,
				UserName:          m.User.Username,
				UserID:            userID,
				Discriminator:     discriminator,
				DiscriminatorOnly: true,
				NamesToCheck:      namesToCheck,
			})
		}

		officialUsers = newOfficialUsers
		log.Printf("We are now monitoring %d official users for imposters\n", len(officialUsers))

		select {
		case <-stopUpdateOfficialUsers:
			return
		case <-time.After(time.Minute * 5):
		}
	}
}

func scanForImpostors(m *discordgo.Member) {
	userID := m.User.ID
	log.Printf("Checking if %s#%s is an impostor\n", m.User.Username, m.User.Discriminator)
	for _, o := range officialUsers {

		nameMatch := false
		for _, n := range o.NamesToCheck {
			if strings.ToLower(m.Nick) == strings.ToLower(n) || strings.ToLower(m.User.Username) == strings.ToLower(n) {
				log.Printf("User %s (%s#%s) matches official name %s (%s#%s)\n", m.Nick, m.User.Username, m.User.Discriminator, o.Nick, o.UserName, o.Discriminator)
				nameMatch = true
				break
			}
		}

		if nameMatch && userID != o.UserID {
			if !o.DiscriminatorOnly || m.User.Discriminator == o.Discriminator {
				alreadyNotified := false
				for _, i := range notifiedImpostors {
					if i == userID {
						alreadyNotified = true
					}
				}

				if !alreadyNotified {
					impostorNotificationChannel <- m
					notifiedImpostors = append(notifiedImpostors, userID)
				}
			}
		}
	}
}

func scanMembers() {
	for {
		log.Printf("Scanning for failed captchas...")
		for _, m := range memberList {
			joined, _ := m.JoinedAt.Parse()
			if joined.Before(time.Now().Add(time.Hour*-1)) && len(m.Roles) == 0 {
				kickChannel <- Kick{Member: m, Reason: "Did not complete captcha within 60 minutes"}
			}
		}

		select {
		case <-stopScanMembers:
			return
		case <-time.After(time.Minute * 1):
		}
	}

}

func removeMember(discord *discordgo.Session, memberRemove *discordgo.GuildMemberRemove) {
	removeIdx := -1
	for i := range memberList {
		if memberList[i].User.ID == memberRemove.Member.User.ID {
			removeIdx = i
		}
	}
	if removeIdx > -1 {
		memberListLock.Lock()
		memberList = append(memberList[:removeIdx], memberList[removeIdx+1:]...)
		memberListLock.Unlock()
		log.Printf("Removed member %s (%s) from index %d\n", memberRemove.Member.User.Username, memberRemove.Member.User.ID, removeIdx)
	}
	log.Printf("We now have %d known members\n", len(memberList))
}

func addMember(discord *discordgo.Session, memberAdd *discordgo.GuildMemberAdd) {
	memberListLock.Lock()
	memberList = append(memberList, memberAdd.Member)
	memberListLock.Unlock()

	scanForImpostors(memberAdd.Member)

	log.Printf("Appended member %s (%s)\n", memberAdd.Member.User.Username, memberAdd.Member.User.ID)
	log.Printf("We now have %d known members\n", len(memberList))
}

func updateMember(discord *discordgo.Session, memberUpdate *discordgo.GuildMemberUpdate) {
	updateIdx := -1
	for i := range memberList {
		if memberList[i].User.ID == memberUpdate.Member.User.ID {
			updateIdx = i
		}
	}
	if updateIdx != -1 {
		memberList[updateIdx] = memberUpdate.Member
		log.Printf("Updated member %s (%s)\n", memberUpdate.Member.User.Username, memberUpdate.Member.User.ID)
	}
	scanForImpostors(memberUpdate.Member)
	log.Printf("We now have %d known members\n", len(memberList))
}

func scanMemberList(discord *discordgo.Session, ready *discordgo.Ready) {
	lastID := "0"
	for {
		members, err := discord.GuildMembers(vertcoinGuildID, lastID, 1000)
		if err != nil {
			break
		}
		memberListLock.Lock()
		memberList = append(memberList, members...)
		memberListLock.Unlock()
		if len(members) < 1000 {
			break
		}
		lastID = members[len(members)-1].User.ID
		log.Printf("Busy scanning server members. Currently at %d known members\n", len(memberList))

		<-time.After(time.Millisecond * 500)
	}

	log.Printf("Done scanning server members. We now have %d known members\n", len(memberList))

	discord.AddHandler(addMember)
	discord.AddHandler(removeMember)
	discord.AddHandler(updateMember)

	for _, m := range memberList {
		scanForImpostors(m)
	}

	go updateProtectedUsers()
	go scanMembers()
}

func errCheck(msg string, err error) {
	if err != nil {
		log.Printf("%s: %+v", msg, err)
		panic(err)
	}
}

func commandHandler(discord *discordgo.Session, message *discordgo.MessageCreate) {
	user := message.Author
	if user.ID == botID || user.Bot {
		//Do nothing because the bot is talking
		return
	}

	if message.GuildID != vertcoinGuildID {
		return
	}

	if strings.HasPrefix(message.Message.Content, "vb!") {
		msg := message.Message.Content[3:]

		discord.ChannelMessageSend(message.ChannelID, fmt.Sprintf("You said: %s", msg))
	}
}
