package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/lfconfig"
	"github.com/ergomake/layerform/internal/validation"
)

func init() {
	cloudCreateUserCmd.Flags().StringP("name", "n", "", "name of the new user")
	cloudCreateUserCmd.Flags().StringP("email", "e", "", "email of the new user")
	cloudCreateUserCmd.MarkFlagRequired("email")

	cloudCmd.AddCommand(cloudCreateUserCmd)
}

var cloudCreateUserCmd = &cobra.Command{
	Use:   "create-user",
	Short: "Creates a new user in layerform cloud",
	Long: `Creates a new user in layerform cloud.

  The password will be printed to stdout, e-mail must be unique.`,
	Example: `# Set a context of type local named local-example
layerform cloud create-user --name "John Doe" --email "john@doe.com"`,
	Run: func(cmd *cobra.Command, _ []string) {
		logger := hclog.Default()
		logLevel := hclog.LevelFromString(os.Getenv("LF_LOG"))
		if logLevel != hclog.NoLevel {
			logger.SetLevel(logLevel)
		}
		ctx := hclog.WithContext(context.Background(), logger)

		cfg, err := lfconfig.Load("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to load config"))
			os.Exit(1)
			return
		}

		currentCfgCtx := cfg.GetCurrent()
		if currentCfgCtx.Type != "cloud" {
			fmt.Fprintf(
				os.Stderr,
				"This command only works if the current context is of type \"cloud\" but current has type \"%s\".\n",
				currentCfgCtx.Type,
			)
			os.Exit(1)
			return
		}

		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get --name flag, this is a bug in layerform"))
			os.Exit(1)
			return
		}
		name = strings.TrimSpace(name)

		email, err := cmd.Flags().GetString("email")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get --email flag, this is a bug in layerform"))
			os.Exit(1)
			return
		}
		email = strings.TrimSpace(email)

		if !validation.IsValidEmail(email) {
			fmt.Fprintf(os.Stderr, "Invalid email \"%s\"\n", email)
			os.Exit(1)
			return

		}

		cloudClient, err := cfg.GetCloudClient(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to get cloud client"))
			os.Exit(1)
		}

		payload, err := json.Marshal(map[string]string{"name": name, "email": email})
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail marshal create user payload to json"))
			os.Exit(1)
			return
		}

		req, err := cloudClient.NewRequest(ctx, "POST", "/v1/users", bytes.NewReader(payload))
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to marshal create user payload to json"))
			os.Exit(1)
			return
		}

		res, err := cloudClient.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to perform http request to cloud backend"))
			os.Exit(1)
			return
		}
		defer res.Body.Close()

		if res.StatusCode == http.StatusConflict {
			fmt.Fprintf(os.Stderr, "User with email %s already exists.\n", email)
			os.Exit(1)
			return
		}

		var body struct {
			Password string `json:"password"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail decode create user JSON response"))
			os.Exit(1)
			return

		}

		identifier := name
		if identifier == "" {
			identifier = email
		}
		fmt.Fprintf(os.Stdout, "User %s created successfully.\nPassword: %s\n", identifier, body.Password)
	},
	SilenceErrors: true,
}
