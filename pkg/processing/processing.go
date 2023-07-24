package processing

import (
	"pagerdutyBot/pkg/controller"
	"pagerdutyBot/pkg/model"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

type Processor struct {
	Pd *controller.PagerDutyConnection
	Sc *controller.SlackConnection
}

func NewProcessor(tokens model.Tokens) (*Processor, error) {
	pd := controller.NewPagerDutyConnection(tokens)
	sc, err := controller.NewSlackConnection(tokens)
	if err != nil {
		return nil, err
	}

	return &Processor{
		Pd: pd,
		Sc: sc,
	}, nil
}

// Read the PD schedule resolve it back to emailID
// These emails IDs are then converted to slack UserIDs.
// func (p *Processor) Schedules(userGroups []model.UserGroupConfig) error {
// 	for _, userGroup := range userGroups {
// 		var userEmails []string
// 		for _, schedule := range userGroup.ScheduleConfig.PagerDutySchedules {
// 			emailsOfSchedule, err := p.Pd.GetOnCallUsers(schedule, userGroup.ScheduleConfig.LookbackPeriod)
// 			if err != nil {
// 				log.Printf("Error getting oncall users for group %s: %v", userGroup.SlackUserGroupName, err)
// 				continue
// 			}
// 			userEmails = append(userEmails, emailsOfSchedule...)
// 		}

// 		log.Debug().Msg("Getting UserID from email")
// 		userIDs, err := p.Sc.GetSlackIDsFromEmails(userEmails)
// 		if err != nil {
// 			log.Printf("Error getting slack user IDs for group %s: %v", userGroup.SlackUserGroupName, err)
// 			continue
// 		}

// 		log.Debug().Msg("Creating usergroup")
// 		ug, err := p.Sc.CreateOrGetUserGroup(userGroup.SlackUserGroupName)
// 		if err != nil {
// 			log.Printf("Error creating or updating user group %v: %v", userGroup.SlackUserGroupName, err)
// 		}

// 		log.Debug().Msg("Getting Current usergroup members")
// 		currentMembers, err := p.Sc.Socket.GetUserGroupMembers(ug.ID)
// 		if err != nil {
// 			log.Printf("Error getting slack user group members %s: %v", userGroup.SlackUserGroupName, err)
// 			continue
// 		}
// 		log.Debug().Msg("Current Members :", currentMembers)

// 		if !slicesEqual(currentMembers, userIDs) {
// 			log.Printf("User group updated needed : %v", userGroup.SlackUserGroupName)
// 			_, err := p.Sc.Socket.Client.UpdateUserGroupMembers(ug.ID, strings.Join(userIDs, ","))
// 			if err != nil {
// 				log.Printf("Error updating slack user group members %s: %v\n%v", userGroup.SlackUserGroupName, err, strings.Join(userIDs, ","))
// 			}
// 		} else {
// 			log.Printf("User Group %v, already in sync", userGroup.SlackUserGroupName)
// 		}

// 	}
// 	return nil
// }

func (p *Processor) Schedules(userGroups []model.UserGroupConfig) error {
	processUserGroupCh := make(chan model.UserGroupConfig, len(userGroups))
	createSlackUserGroupsCh := make(chan model.UserGroupResult, len(userGroups))
	doneCh := make(chan struct{})

	var wg sync.WaitGroup

	log.Trace().Msg("Starting Go Routines")
	go p.processUserGroupsWorker(&wg, processUserGroupCh, createSlackUserGroupsCh)
	go p.createSlackUserGroupsWorker(&wg, createSlackUserGroupsCh, doneCh)

	log.Trace().Msg("Looping over usergroups")
	for _, userGroup := range userGroups {
		wg.Add(1)
		processUserGroupCh <- userGroup
	}

	close(processUserGroupCh)
	wg.Wait()
	close(createSlackUserGroupsCh)
	<-doneCh

	return nil
}

func (p *Processor) processUserGroupsWorker(wg *sync.WaitGroup, inputCh <-chan model.UserGroupConfig, outputCh chan<- model.UserGroupResult) {

	log.Trace().Msg("Inside Process user group")

	for userGroup := range inputCh {
		log.Debug().Msgf("Processing User Group : %v", userGroup.SlackUserGroupName)
		var userEmails []string
		for _, schedule := range userGroup.ScheduleConfig.PagerDutySchedules {
			emailsOfSchedule, err := p.Pd.GetOnCallUsers(schedule, userGroup.ScheduleConfig.LookbackPeriod)
			if err != nil {
				log.Error().Msgf("Error getting oncall users for group %s: %v", userGroup.SlackUserGroupName, err)
				wg.Done()
				continue
			}
			userEmails = append(userEmails, emailsOfSchedule...)
		}
		log.Debug().Msgf("User Group : %v has emails : %v", userGroup.SlackUserGroupName, userEmails)
		userIDs, err := p.Sc.GetSlackIDsFromEmails(userEmails)
		if err != nil {
			log.Error().Msgf("Error getting slack user IDs for group %s: %v", userGroup.SlackUserGroupName, err)
			wg.Done()
			continue
		}

		result := model.UserGroupResult{
			UserGroupName: userGroup.SlackUserGroupName,
			OnCallUsers:   userIDs,
		}

		log.Debug().Msgf("Message sent to output channel for usergroup creation : %v, with members :%v", userGroup.SlackUserGroupName, userIDs)
		outputCh <- result
	}
}

func (p *Processor) createSlackUserGroupsWorker(wg *sync.WaitGroup, inputCh <-chan model.UserGroupResult, doneCh chan<- struct{}) {

	log.Trace().Msg("Inside create slack user group")

	for schedule := range inputCh {
		log.Debug().Msgf("Input received from output channel for usergroup creation : %v", schedule.UserGroupName)

		userGroup, err := p.Sc.CreateOrGetUserGroup(schedule.UserGroupName)
		if err != nil {
			log.Error().Msgf("Error creating or updating user group %s: %v", schedule.UserGroupName, err)
			continue
		}
		log.Debug().Msgf("Found Usergroup %v , with ID: %v", userGroup.Name, userGroup.ID)

		currentMembers, err := p.Sc.Socket.GetUserGroupMembers(userGroup.ID)
		if err != nil {
			log.Error().Msgf("Error getting slack user group members %s: %v", userGroup.Name, err)
			continue
		}
		log.Debug().Msgf("Current members of Usergroup %v : %v", userGroup.Name, currentMembers)

		if !slicesEqual(currentMembers, schedule.OnCallUsers) {
			log.Info().Msgf("User group updated needed : %v", schedule.UserGroupName)
			_, err := p.Sc.Socket.Client.UpdateUserGroupMembers(userGroup.ID, strings.Join(schedule.OnCallUsers, ","))
			if err != nil {
				log.Error().Msgf("Error updating slack user group members %s: %v", userGroup.Name, err)
			}
			log.Info().Msgf("Updated members of Usergroup %v : %v", userGroup.Name, currentMembers)
		} else {
			log.Info().Msgf("User Group is up-to-date : %v", userGroup.Name)
		}
		wg.Done() //Indicate Processing of schedule is done
	}

	log.Debug().Msg("Processing schedules from output channel Completed. Sending DONE")

	doneCh <- struct{}{}
}
