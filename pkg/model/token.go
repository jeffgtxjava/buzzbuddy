package model

type Tokens struct {
	SlackAppToken     string `yaml:"slack_app_token"`
	SlackBotToken     string `yaml:"slack_bot_token"`
	PagerDutyAPIToken string `yaml:"pagerduty_api_token"`
}
