package bot

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// stockHandler will be called every time a new
// message is created on any channel that the autenticated bot has access to.
func (b *quartermasterBot) stockHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!stock" {
		allContracts, err := b.loadContracts()

		if err != nil {
			b.log.Errorw("error loading ESI contracts", "error", err)
			b.sendError(err, m)
			return
		}
		corporationContracts, allianceContracts := b.filterAndGroupContracts(
			allContracts,
			"outstanding",
			true,
		)
		gotCorporationDoctrines := doctrinesAvailable(corporationContracts)
		gotAllianceDoctrines := doctrinesAvailable(allianceContracts)
		_, err = b.discord.ChannelMessageSendEmbed(
			m.ChannelID,
			stockMessage(gotCorporationDoctrines, gotAllianceDoctrines),
		)
		if err != nil {
			b.log.Errorw("error sending message for !stock", "error", err)
			return
		}
		return
	}
}

func stockMessage(corporationDoctrines, allianceDoctrines map[string]int) *discordgo.MessageEmbed {
	var (
		namesCorporation, namesAlliance []string // used for sorting by name
		partsCorporation, partsAlliance []string
	)

	for haveDoctrine := range corporationDoctrines {
		namesCorporation = append(namesCorporation, haveDoctrine)
	}
	sort.Strings(namesCorporation)

	for haveDoctrine := range allianceDoctrines {
		namesAlliance = append(namesAlliance, haveDoctrine)
	}
	sort.Strings(namesAlliance)

	for _, name := range namesCorporation {
		partsCorporation = append(partsCorporation, fmt.Sprintf("%d %s", corporationDoctrines[name], name))
	}

	for _, name := range namesAlliance {
		partsAlliance = append(partsAlliance, fmt.Sprintf("%d %s", allianceDoctrines[name], name))
	}

	msg := fmt.Sprintf(
		"**Alliance contracts**\n```\n%s\n```\n**Corporation contracts**\n```\n%s\n```",
		strings.Join(partsAlliance, "\n"),
		strings.Join(partsCorporation, "\n"),
	)

	return &discordgo.MessageEmbed{
		Title: "Have on contract",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://i.imgur.com/ZwUn8DI.jpg",
		},
		Color:       0x00ff00,
		Description: msg,
		Timestamp:   time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
	}
}
