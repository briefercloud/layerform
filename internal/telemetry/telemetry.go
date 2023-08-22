package telemetry

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/posthog/posthog-go"
)

type client struct {
	posthog.Client
}

var c *client

func Init() {
	if !isEnabled() {
		return
	}

	ph, err := posthog.NewWithConfig(
		"phc_AT86CaGmI1DoCIFHOPUwIhN3OQ8uo6nT2AhpQb2GvMC",
		posthog.Config{
			Endpoint: "https://app.posthog.com",
		},
	)

	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "fail to init telemetry"))
	}

	c = &client{ph}
}

func isEnabled() bool {
	env := strings.ToLower(os.Getenv("LF_TELEMETRY_DISABLED"))
	return env != "1" && env != "yes" && env != "true"
}

func Close() {
	if c == nil || !isEnabled() {
		return
	}

	err := c.Close()
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "fail to close telemetry"))
	}
}

type Event string

const EventRunCommand Event = "user run command"

func Push(event Event, properties map[string]interface{}) {
	if c == nil || !isEnabled() {
		return
	}

	err := c.Enqueue(posthog.Capture{
		DistinctId: getDistinctId(),
		Event:      string(event),
		Properties: properties,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "fail to enqueue telemetry"))
	}
}

func getDistinctId() string {
	id := uuid.NewString()

	homedir, err := os.UserHomeDir()
	if err != nil {
		return id
	}

	lfDir := path.Join(homedir, ".layerform")
	err = os.MkdirAll(lfDir, os.ModePerm)
	if err != nil {
		return id
	}

	fpath := path.Join(lfDir, "userid")

	bs, err := os.ReadFile(fpath)
	if err != nil {
		os.WriteFile(fpath, []byte(id), 0644)
		return id
	}

	prevId, err := uuid.ParseBytes(bs)
	if err != nil {
		os.WriteFile(fpath, []byte(id), 0644)
		return id
	}

	return prevId.String()
}
