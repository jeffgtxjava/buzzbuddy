// pagerduty.go
package controller

import (
	"time"

	"github.com/rs/zerolog/log"

	"pagerdutyBot/pkg/model"

	"github.com/PagerDuty/go-pagerduty"
)

type PagerDutyConnection struct {
	client *pagerduty.Client
}

func NewPagerDutyConnection(tokens model.Tokens) *PagerDutyConnection {
	client := pagerduty.NewClient(tokens.PagerDutyAPIToken)
	return &PagerDutyConnection{client}
}

func (pdConn *PagerDutyConnection) GetOnCallUsers(scheduleID string, lookbackPeriod time.Duration) ([]string, error) {
	log.Trace().Msg("Inside GetOncallUsers")
	var userEmails []string
	since := time.Now().Format(time.RFC3339)
	until := time.Now().Add(lookbackPeriod).Format(time.RFC3339)

	options := pagerduty.ListOnCallUsersOptions{
		Since: since,
		Until: until,
	}

	users, err := pdConn.client.ListOnCallUsers(scheduleID, options)
	if err != nil {
		return nil, err
	}

	for _, oncall := range users {
		userEmails = append(userEmails, oncall.Email)
	}

	log.Debug().Msgf("Users for Schedule(%v) : %v", scheduleID, userEmails)
	return userEmails, nil
}
