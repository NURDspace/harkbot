package main

import (
	"flag"
	"harkbot/lib/broker"
	"harkbot/lib/common"
	"harkbot/lib/config"
	"harkbot/lib/ircbot"
	"harkbot/lib/ircstate"
)

func main() {
	var (
		cfgFile           = flag.String("cfgfile", config.DefaultCfgFile, config.HelpCfgFile)
		flagServer        = flag.String("server", config.DefaultServer, config.HelpServer)
		flagUseTls        = flag.Bool("usetls", config.DefaultUseTls, config.HelpUseTls)
		flagVerifyTls     = flag.Bool("verifytls", config.DefaultVerifyTls, config.HelpVerifyTls)
		flagChannel       = flag.String("channel", config.DefaultChannel, config.HelpChannel)
		flagNickname      = flag.String("nickname", config.DefaultNickname, config.HelpNickname)
		flagRealname      = flag.String("realname", config.DefaultRealname, config.HelpRealName)
		flagOwner         = flag.String("owner", config.DefaultOwner, config.HelpOwner)
		flagCommandPrefix = flag.String("commandprefix", config.DefaultCommandPrefix, config.HelpCommandPrefix)
		flagDbPath        = flag.String("dbpath", config.DefaultDbPath, config.HelpDbPath)
		flagPluginPath    = flag.String("pluginpath", config.DefaultPluginPath, config.HelpPluginPath)
	)

	flag.Parse()

	settings := config.Load(*cfgFile, &config.Settings{
		Server:        *flagServer,
		UseTls:        *flagUseTls,
		VerifyTls:     *flagVerifyTls,
		Channel:       *flagChannel,
		Nickname:      *flagNickname,
		Realname:      *flagRealname,
		Owner:         *flagOwner,
		CommandPrefix: *flagCommandPrefix,
		DbPath:        *flagDbPath,
		PluginPath:    *flagPluginPath,
	})

	unfilteredChan := make(chan common.RawMessage, broker.MaxInFlight)
	commandChan := make(chan common.CommandMessage, broker.MaxInFlight)

	privmsgChan := make(chan common.FromMessage, broker.MaxInFlight)
	noticeChan := make(chan common.FromMessage, broker.MaxInFlight)
	topicChan := make(chan common.TopicMessage, broker.MaxInFlight)

	mqtt := broker.New(&broker.Params{
		Server:         "localhost:1883",
		ClientId:       settings.Nickname,
		UnfilteredChan: unfilteredChan,
		CommandChan:    commandChan,
		PrivmsgChan:    privmsgChan,
		NoticeChan:     noticeChan,
		TopicChan:      topicChan,
	})

	state, err := ircstate.New(settings.DbPath)
	defer state.Close()
	if err != nil {
		panic(err)
	}

	bot, err := ircbot.New(&ircbot.Params{
		Server:         settings.Server,
		UseTLS:         settings.UseTls,
		VerifyTLS:      settings.VerifyTls,
		Channel:        settings.Channel,
		Nickname:       settings.Nickname,
		Realname:       settings.Realname,
		CommandPrefix:  settings.CommandPrefix,
		Owner:          settings.Owner,
		UnfilteredChan: unfilteredChan,
		CommandChan:    commandChan,
		PrivmsgChan:    privmsgChan,
		NoticeChan:     noticeChan,
		TopicChan:      topicChan,
	}, state, mqtt)
	if err != nil {
		panic(err)
	}

	bot.Run()
}
