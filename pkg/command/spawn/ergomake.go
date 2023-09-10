package spawn

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
	"github.com/ergomake/layerform/pkg/layerinstances"
)

type ergomakeSpawnCommand struct {
	baseURL          string
	instancesBackend layerinstances.Backend
}

var _ Spawn = &ergomakeSpawnCommand{}

func NewErgomake(baseURL string) *ergomakeSpawnCommand {
	instancesBackend := layerinstances.NewErgomake(baseURL)

	return &ergomakeSpawnCommand{baseURL, instancesBackend}
}

func (e *ergomakeSpawnCommand) Run(
	ctx context.Context,
	definitionName, instanceName string,
	dependenciesInstance map[string]string,
	vars []string,
) error {
	logger := hclog.FromContext(ctx)
	logger.Debug("Spawning instance remotely")

	_, err := e.instancesBackend.GetInstance(ctx, definitionName, instanceName)
	if err == nil {
		return errors.Errorf("layer %s already spawned with name %s", definitionName, instanceName)
	}
	if !errors.Is(err, layerinstances.ErrInstanceNotFound) {
		return errors.Wrap(err, "fail to get instance")
	}

	url := fmt.Sprintf("%s/v1/definitions/%s/instances/%s/spawn", e.baseURL, definitionName, instanceName)
	dataBytes, err := json.Marshal(nil)
	if err != nil {
		return errors.Wrap(err, "fail to marshal instance to json")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataBytes))
	if err != nil {
		return errors.Wrap(err, "fail to create http request to ergomake backend")
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		return errors.Wrap(err, "fail to perform http request to ergomake backend")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("HTTP request to %s failed with status code %d", url, resp.StatusCode)
	}

	// TODO: improve the loading experience, make it similar to the
	// localSpawnCommand to accomplish that we need to communicate with
	// layerformcloud to get spawn logs or something like that
	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()
	s := sm.AddSpinner(fmt.Sprintf("Spawning instance \"%s\" of layer \"%s\" remotely", instanceName, definitionName))

	time.Sleep(time.Second * 2)
	for {
		instance, err := e.instancesBackend.GetInstance(ctx, definitionName, instanceName)
		if err != nil {
			s.Error()
			sm.Stop()
			return errors.Wrap(err, "fail to get instance to check spawning status")
		}

		switch instance.Status {
		case data.LayerInstanceStatusSpawning:
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
