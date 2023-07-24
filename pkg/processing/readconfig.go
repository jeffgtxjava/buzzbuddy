package processing

import (
	"io/ioutil"
	"pagerdutyBot/pkg/model"

	"github.com/rs/zerolog/log"

	"gopkg.in/yaml.v2"
)

func ReadConfig(filePath string) []model.UserGroupConfig {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal().Err(err)
	}

	var config []model.UserGroupConfig
	err = yaml.UnmarshalStrict(content, &config)
	if err != nil {
		log.Fatal().Err(err)
	}

	if hasDuplicateUserGroups(config) {
		log.Fatal().Msg("Fix the config and reload the program")
	}

	return config
}

func hasDuplicateUserGroups(userGroups []model.UserGroupConfig) bool {
	seen := make(map[string]bool)
	duplicate := false
	for _, group := range userGroups {
		if _, ok := seen[group.SlackUserGroupName]; ok {
			log.Error().Msgf("Error: Duplicate user group name found: %s", group.SlackUserGroupName)
			duplicate = true
		} else {
			seen[group.SlackUserGroupName] = true
		}
	}
	return duplicate
}

func ReadTokens(filePath string) model.Tokens {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatal().Err(err)
	}

	var tokens model.Tokens
	err = yaml.Unmarshal(content, &tokens)
	if err != nil {
		log.Fatal().Err(err)
	}

	return tokens
}
