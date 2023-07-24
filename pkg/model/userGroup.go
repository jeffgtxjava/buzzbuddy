package model

import (
	"time"
)

// UserGroupResult holds the results of the processing for a user group
type UserGroupResult struct {
	UserGroupName string
	OnCallUsers   []string
	Error         error
}

//Parse the input YAML configuration file
type UserGroupConfig struct {
	SlackUserGroupName string `yaml:"slack_user_group_name"`
	ScheduleConfig     struct {
		PagerDutySchedules []string      `yaml:"pager_duty_schedules"`
		LookbackPeriod     time.Duration `yaml:"lookback_period"`
	} `yaml:"schedule_config"`
}
