package cmd

import (
	"fmt"
	"time"

	"github.com/lunemec/eve-quartermaster/pkg/bot"
	"github.com/lunemec/eve-quartermaster/pkg/repository"
	"github.com/lunemec/eve-quartermaster/pkg/token"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the discord bot",
	Run:   runBot,
}

var (
	checkInterval  time.Duration
	notifyInterval time.Duration

	corporationID int32
	allianceID    int32

	discordChannelID string
	discordAuthToken string

	repositoryFile string
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&authfile, "auth_file", "a", "auth.bin", "path to file where to save authentication data")
	runCmd.Flags().StringVarP(&sessionKey, "session_key", "s", "", "session key, use random string")
	runCmd.Flags().StringVar(&eveClientID, "eve_client_id", "", "EVE APP client id")
	runCmd.Flags().StringVar(&eveSSOSecret, "eve_sso_secret", "", "EVE APP SSO secret")
	runCmd.Flags().StringVar(&discordChannelID, "discord_channel_id", "", "ID of discord channel")
	runCmd.Flags().StringVar(&discordAuthToken, "discord_auth_token", "", "Auth token for discord")
	runCmd.Flags().Int32Var(&corporationID, "corporation_id", 0, "Corporation ID for which to list contracts")
	runCmd.Flags().Int32Var(&allianceID, "alliance_id", 0, "Alliance ID for which to list contracts")
	runCmd.Flags().DurationVar(&checkInterval, "check_interval", 30*time.Minute, "how often to check EVE ESI API (default 30min)")
	runCmd.Flags().DurationVar(&notifyInterval, "notify_interval", 24*time.Hour, "how often to spam Discord (default 24H)")
	runCmd.Flags().StringVar(&repositoryFile, "repository_file", "repository.json", "path to repository json to save require_stock data (default repository.json)")

	must(runCmd.MarkFlagRequired("session_key"))
	must(runCmd.MarkFlagRequired("eve_client_id"))
	must(runCmd.MarkFlagRequired("eve_sso_secret"))
	must(runCmd.MarkFlagRequired("discord_channel_id"))
	must(runCmd.MarkFlagRequired("discord_auth_token"))
}

func runBot(cmd *cobra.Command, args []string) {
	fastLog, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("error inicializing logger: %s", err))
	}
	log := fastLog.Sugar()

	client := httpClient()

	tokenStorage := token.NewFileStorage(authfile)
	tokenSource := token.NewSource(
		log,
		client,
		tokenStorage,
		[]byte(sessionKey),
		eveClientID,
		eveSSOSecret,
		eveCallbackURL,
		eveScopes,
	)

	discord, err := discordgo.New("Bot " + discordAuthToken)
	if err != nil {
		panic(fmt.Sprintf("error inicializing discord client: %+v", err))
	}

	repository, err := repository.NewJSONRepository(repositoryFile)
	if err != nil {
		panic(fmt.Sprintf("error inicializing repository file: %+v", err))
	}

	bot := bot.NewQuartermasterBot(
		log,
		client,
		tokenSource,
		discord,
		discordChannelID,
		corporationID,
		allianceID,
		repository,
		checkInterval,
		notifyInterval,
	)
	err = bot.Bot()
	// systemd handles reload, so we can panic on error.
	if err != nil {
		panic(err)
	}
}
