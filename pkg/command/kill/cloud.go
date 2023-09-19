package kill

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

	"github.com/ergomake/layerform/internal/cloud"
	"github.com/ergomake/layerform/pkg/data"
	"github.com/ergomake/layerform/pkg/layerdefinitions"
	"github.com/ergomake/layerform/pkg/layerinstances"
)

type cloudKillCommand struct {
	client             *cloud.HTTPClient
	definitionsBackend layerdefinitions.Backend
	instancesBackend   layerinstances.Backend
}

var _ Kill = &cloudKillCommand{}

func NewCloud(client *cloud.HTTPClient) *cloudKillCommand {
	definitionsBackend := layerdefinitions.NewCloud(client)
	instancesBackend := layerinstances.NewCloud(client)

	return &cloudKillCommand{client, definitionsBackend, instancesBackend}
}

func (e *cloudKillCommand) Run(
	ctx context.Context,
	definitionName, instanceName string,
	autoApprove bool,
	vars []string,
	force bool,
) error {
	logger := hclog.FromContext(ctx)
	logger.Debug("Killing instance remotely")

	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()
	s := sm.AddSpinner(
		fmt.Sprintf(
			"Preparing to kill instance \"%s\" of layer \"%s\"",
			instanceName,
			definitionName,
		),
	)

	definition, err := e.definitionsBackend.GetLayer(ctx, definitionName)
	if err != nil {
		return errors.Wrap(err, "fail to get layer")
	}

	if definition == nil {
		return errors.New("layer not found")
	}

	_, err = e.instancesBackend.GetInstance(ctx, definition.Name, instanceName)
	if err != nil {
		if errors.Is(err, layerinstances.ErrInstanceNotFound) {
			return errors.Errorf(
				"instance %s not found for layer %s",
				instanceName,
				definition.Name,
			)
		}

		return errors.Wrap(err, "fail to get layer instance")
	}

	hasDependants, err := HasDependants(
		ctx,
		e.instancesBackend,
		e.definitionsBackend,
		definitionName,
		instanceName,
	)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to check if layer has dependants")
	}

	if hasDependants {
		s.Error()
		sm.Stop()
		return errors.New("can't kill this layer because other layers depend on it")
	}

	url := fmt.Sprintf("/v1/definitions/%s/instances/%s/kill", definitionName, instanceName)
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

	req, err := e.client.NewRequest(ctx, "POST", url, bytes.NewBuffer(dataBytes))
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to create http request to cloud backend")
	}

	req.SetHeader("Content-Type", "application/json")
	resp, err := e.client.Do(req)
	if err != nil {
		s.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to perform http request to cloud backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.Error()
		sm.Stop()
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	s.Complete()

	// TODO: improve the loading experience, make it similar to the
	// localKillCommand to accomplish that we need to communicate with
	// layerformcloud to get spawn logs or something like that
	s = sm.AddSpinner(fmt.Sprintf("Killing instance \"%s\" of layer \"%s\" remotely", instanceName, definitionName))

	time.Sleep(time.Second * 2)
	for {
		instance, err := e.instancesBackend.GetInstance(ctx, definitionName, instanceName)
		if errors.Is(err, layerinstances.ErrInstanceNotFound) {
			s.Complete()
			sm.Stop()
			return nil
		}

		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "fail to get instance to check killing status")
		}

		switch instance.Status {
		case data.LayerInstanceStatusKilling:
			time.Sleep(time.Second * 2)
			continue
		case data.LayerInstanceStatusFaulty:
			s.Error()
			sm.Stop()
			return errors.Errorf("fail to kill instance %s of definition %s", instanceName, definitionName)
		default:
			s.Error()
			sm.Stop()
			return errors.Errorf("instance entered a unexpected status of %s", string(instance.Status))
		}
	}
}
