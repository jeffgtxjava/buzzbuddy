// slack.go

package controller

import (
	"fmt"
	slogger "log"
	"os"
	"strings"

	"pagerdutyBot/pkg/model"

	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/socketmode"
)

type SlackConnection struct {
	Client     *slack.Client
	Socket     *socketmode.Client
	users      []slack.User
	userGroups []slack.UserGroup
}

func NewSlackConnection(tokens model.Tokens) (*SlackConnection, error) {
	appLevelToken := tokens.SlackAppToken
	botUserLevelToken := tokens.SlackBotToken

	if !strings.HasPrefix(appLevelToken, "xapp-") {
		return nil, fmt.Errorf("SLACK_APP_TOKEN must have the prefix \"xapp-\"")
	}

	if !strings.HasPrefix(botUserLevelToken, "xoxb-") {
		return nil, fmt.Errorf("SLACK_BOT_TOKEN must have the prefix \"xoxb-\"")
	}

	client := slack.New(
		botUserLevelToken,
		slack.OptionAppLevelToken(appLevelToken),
		slack.OptionLog(slogger.New(os.Stdout, "api: ", slogger.Lshortfile|slogger.LstdFlags)),
		slack.OptionDebug(true),
	)

	socketmodeClient := socketmode.New(
		client,
		socketmode.OptionDebug(true),
		socketmode.OptionLog(slogger.New(os.Stdout, "socketmode: ", slogger.Lshortfile|slogger.LstdFlags)),
	)

	log.Trace().Msg("Reading User Configs")
	userGroups, err := socketmodeClient.Client.GetUserGroups()
	if err != nil {
		return nil, err
	}

	log.Trace().Msg("Creating Clients")
	return &SlackConnection{
		Client:     client,
		Socket:     socketmodeClient,
		users:      []slack.User{},
		userGroups: userGroups,
	}, nil
}

func (sc *SlackConnection) GetSlackIDsFromEmails(emails []string) ([]string, error) {
	log.Trace().Msg("Inside GetSlackIDsFromEmails")
	var results []string
	for _, email := range emails {
		ID := sc.findUserIDByEmail(email)
		if ID == nil {
			return nil, fmt.Errorf("could not find slack user with email: %s", email)
		}
		results = append(results, *ID)
	}
	return results, nil
}

func (sc *SlackConnection) CreateOrGetUserGroup(slackGroupName string) (*slack.UserGroup, error) {
	log.Trace().Msg("Inside CreateOrGetUserGroup")
	group := sc.findUserGroupByName(slackGroupName)
	if group != nil {
		return group, nil
	}

	g, err := sc.Client.CreateUserGroup(slack.UserGroup{
		Name:   slackGroupName,
		Handle: slackGroupName,
	})

	if err != nil {
		log.Error().Msgf("Error while creating group %v: %v", slackGroupName, err)
		return nil, err
	}

	return &g, err
}

func (sc *SlackConnection) findUserIDByEmail(email string) *string {
	suser, err := sc.Client.GetUserByEmail(email)
	if err != nil {
		log.Error().Msgf("Error finding the user %v: %v", email, err)
		return nil
	}
	return &suser.ID
}

func (sc *SlackConnection) findUserGroupByName(name string) *slack.UserGroup {
	for _, g := range sc.userGroups {
		if strings.EqualFold(name, g.Name) {
			return &g
		}
	}
	return nil
}
