package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ergomake/layerform/internal/lfconfig"
)

func init() {
	configSetContextCmd.Flags().StringP("type", "t", "local", "type of the context entry, must be \"local\", \"s3\" or \"cloud\"")
	configSetContextCmd.Flags().String("dir", "", "directory to store definitions and instances, required when type is \"local\"")
	configSetContextCmd.Flags().String("bucket", "", "bucket to store definitions and instances, required when type is \"s3\"")
	configSetContextCmd.Flags().String("region", "", "region where bucket is located, required when type is \"s3\"")
	configSetContextCmd.Flags().String("url", "", "url of layerform cloud, required when type is \"cloud\"")
	configSetContextCmd.Flags().String("email", "", "email of layerform cloud user, required when type is \"cloud\"")
	configSetContextCmd.Flags().String("password", "", "password of layerform cloud user, required when type is \"cloud\"")
	configSetContextCmd.Flags().SortFlags = false

	configCmd.AddCommand(configSetContextCmd)
}

var configSetContextCmd = &cobra.Command{
	Use:   "set-context <name>",
	Short: "Set a context entry in layerform config file",
	Long: `Set a context entry in layerform config file.

  Specifying a name that already exists will update that context values unless the type is different.`,
	Example: `# Set a context of type local named local-example
layerform config set-context local-example -t local --dir example-dir

# Set a context of type s3 named s3-example
layerform config set-context s3-example -t s3 --bucket example-bucket --region us-east-1

# Set a context of type cloud named cloud-example
layerform config set-context cloud-example -t cloud --url https://example.layerform.dev --email foo@example.com --password secretpass`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		t, _ := cmd.Flags().GetString("type")
		configCtx := lfconfig.ConfigContext{Type: t}
		switch configCtx.Type {
		case "local":
			dir, _ := cmd.Flags().GetString("dir")
			configCtx.Type = t
			configCtx.Dir = strings.TrimSpace(dir)
		case "s3":
			bucket, _ := cmd.Flags().GetString("bucket")
			region, _ := cmd.Flags().GetString("region")
			configCtx.Bucket = strings.TrimSpace(bucket)
			configCtx.Region = strings.TrimSpace(region)
		case "cloud":
			url, _ := cmd.Flags().GetString("url")
			email, _ := cmd.Flags().GetString("email")
			password, _ := cmd.Flags().GetString("password")
			configCtx.URL = strings.TrimSpace(url)
			configCtx.Email = strings.TrimSpace(email)
			configCtx.Password = strings.TrimSpace(password)
		default:
			fmt.Fprintf(os.Stderr, "invalid type %s\n", configCtx.Type)
			os.Exit(1)
		}

		err := lfconfig.Validate(configCtx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "invalid context configuration"))
			os.Exit(1)
		}

		cfg, err := lfconfig.Load("")
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to open config file"))
			os.Exit(1)
		}

		action := "modified"
		if cfg == nil {
			action = "created"
			cfg, err = lfconfig.Init(name, configCtx, "")
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to initialize empty config"))
				os.Exit(1)
			}
		} else {
			prev, ok := cfg.Contexts[name]
			if !ok {
				action = "created"
			}

			if ok && prev.Type != t {
				fmt.Fprintf(
					os.Stderr,
					"%s context already exists with a different type of %s, context type can't be updated.\n",
					name,
					cfg.GetCurrent().Type,
				)
				os.Exit(1)
			}
			cfg.Contexts[name] = configCtx
		}

		cfg.CurrentContext = name

		err = cfg.Save()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", errors.Wrap(err, "fail to save config file"))
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "Context \"%s\" %s.\n", name, action)
	},
	SilenceErrors: true,
}
