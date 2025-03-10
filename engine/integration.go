package engine

import (
	"github.com/theapemachine/idrinkyourmilkshake/models"
)

/*
Integration is the execution engine for a given API config.
It will execute the API calls in the order specified by the API config.
It will also handle the pagination of the API calls if the API config specifies it.
*/
type Integration struct {
	config models.APIConfig
}

func NewIntegration(config models.APIConfig) *Integration {
	return &Integration{config: config}
}

func (integration *Integration) Execute() (err error) {
	for _, job := range integration.config.Jobs {
		for _, step := range job.Steps {
			if err = integration.executeStep(step); err != nil {
				return err
			}
		}
	}
	return nil
}

func (integration *Integration) executeStep(step models.Step) (err error) {
	_ = step
	return nil
}
