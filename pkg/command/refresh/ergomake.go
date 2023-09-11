package refresh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chelnak/ysmrr"
	"github.com/chelnak/ysmrr/pkg/animations"
	"github.com/chelnak/ysmrr/pkg/colors"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/data"
	"github.com/ergomake/layerform/pkg/layerdefinitions"
	"github.com/ergomake/layerform/pkg/layerinstances"
)

type ergomakeRefreshCommand struct {
	baseURL            string
	instancesBackend   layerinstances.Backend
	definitionsBackend layerdefinitions.Backend
}

var _ Refresh = &ergomakeRefreshCommand{}

func NewErgomake(baseURL string) *ergomakeRefreshCommand {
	instancesBackend := layerinstances.NewErgomake(baseURL)
	definitionsBackend := layerdefinitions.NewErgomake(baseURL)

	return &ergomakeRefreshCommand{baseURL, instancesBackend, definitionsBackend}
}

func (e *ergomakeRefreshCommand) Run(
	ctx context.Context,
	definitionName, instanceName string,
	vars []string,
) error {
	logger := hclog.FromContext(ctx)
	logger.Debug("Refreshing instance remotely")

	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()
	s := sm.AddSpinner(
		fmt.Sprintf(
			"Preparing to refresh instance \"%s\" of layer \"%s\"",
			instanceName,
			definitionName,
		),
	)

	definition, err := e.definitionsBackend.GetLayer(ctx, definitionName)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to get layer definition")
	}

	_, err = e.instancesBackend.GetInstance(ctx, definition.Name, instanceName)
	if err != nil {
		s.Error()
		sm.Stop()

		if errors.Is(err, layerinstances.ErrInstanceNotFound) {
			return errors.Errorf(
				"instance %s not found for layer %s",
				instanceName,
				definition.Name,
			)
		}

		return errors.Wrap(err, "fail to get layer instance")
	}

	s.Complete()
	s = sm.AddSpinner(fmt.Sprintf("Refreshing instance \"%s\" of layer \"%s\" remotely", instanceName, definitionName))

	url := fmt.Sprintf("%s/v1/definitions/%s/instances/%s/refresh", e.baseURL, definitionName, instanceName)
	dataBytes, err := json.Marshal(
		map[string]interface{}{
			"vars": vars,
		},
	)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to marshal instance to json")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataBytes))
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to create http request to ergomake backend")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to perform http request to ergomake backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Error()
		sm.Stop()
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	time.Sleep(time.Second * 2)
	for {
		instance, err := e.instancesBackend.GetInstance(ctx, definitionName, instanceName)
		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "fail to get instance to check spawning status")
		}

		switch instance.Status {
		case data.LayerInstanceStatusRefreshing:
			time.Sleep(time.Second * 2)
			continue
		case data.LayerInstanceStatusFaulty:
			s.Error()
			sm.Stop()
			return errors.Errorf("fail to spawn instance %s of definition %s", instanceName, definitionName)
		case data.LayerInstanceStatusAlive:
			s.Complete()
			sm.Stop()
			return nil
		default:
			s.Error()
			sm.Stop()
			return errors.Errorf("instance entered a unexpected status of %s", string(instance.Status))
		}
	}
}
