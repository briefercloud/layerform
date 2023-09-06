package command

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/chelnak/ysmrr"
	"github.com/chelnak/ysmrr/pkg/animations"
	"github.com/chelnak/ysmrr/pkg/colors"
	"github.com/hashicorp/go-hclog"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/ergomake/layerform/internal/layerdefinitions"
	"github.com/ergomake/layerform/internal/layerfile"
	"github.com/ergomake/layerform/internal/layerinstances"
	"github.com/ergomake/layerform/internal/terraform"
	"github.com/ergomake/layerform/internal/tfclient"
	"github.com/ergomake/layerform/pkg/data"
)

type configureCommand struct {
	definitionssBackend layerdefinitions.Backend
	instancesBackend    layerinstances.Backend
}

func NewConfigure(definitionsBackend layerdefinitions.Backend, instancesBackend layerinstances.Backend) *configureCommand {
	return &configureCommand{definitionsBackend, instancesBackend}
}

func (c *configureCommand) Run(ctx context.Context, fpath string) error {
	logger := hclog.FromContext(ctx)

	sm := ysmrr.NewSpinnerManager(
		ysmrr.WithAnimation(animations.Dots),
		ysmrr.WithSpinnerColor(colors.FgHiBlue),
	)
	sm.Start()

	loadSpinner := sm.AddSpinner(fmt.Sprintf("Loading layer definitions from \"%s\"", fpath))

	layerfile, err := layerfile.FromFile(fpath)
	if err != nil {
		loadSpinner.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to read layerform layers definitions from file")
	}

	ls, err := layerfile.ToLayers()
	if err != nil {
		loadSpinner.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to load layers from layerform layers definitions file")
	}

	if len(ls) == 0 {
		loadSpinner.Error()
		sm.Stop()
		return errors.Errorf("No layers are defined at \"%s\"", fpath)
	}

	loadSpinner.UpdateMessagef(
		"%d %s loaded from \"%s\"",
		len(ls),
		pluralize("layer", len(ls)),
		fpath,
	)
	loadSpinner.Complete()

	tfpath, err := terraform.GetTFPath(ctx)
	if err != nil {
		sm.Stop()
		return errors.Wrap(err, "fail to get terraform path")
	}
	logger.Debug("Using terraform from", "tfpath", tfpath)

	logger.Debug("Creating a temporary work directory")
	workdir, err := os.MkdirTemp("", "")
	if err != nil {
		sm.Stop()
		return errors.Wrap(err, "fail to create work directory")
	}
	defer os.RemoveAll(workdir)

	inMemoryDefinitionsBackend := layerdefinitions.NewInMemoryBackend(ls)
	var wg sync.WaitGroup
	type validationErr struct {
		err         error
		diagnostics string
	}
	errs := make([]validationErr, len(ls))
	for i, l := range ls {
		wg.Add(1)
		go func(i int, l *data.LayerDefinition) {
			defer wg.Done()

			s := sm.AddSpinner(fmt.Sprintf("Validating layer %s", l.Name))

			layerWorkdir := path.Join(workdir, l.Name)

			tfWorkdir, err := writeLayerToWorkdir(ctx, inMemoryDefinitionsBackend, layerWorkdir, l, map[string]string{})
			if err != nil {
				s.Error()
				errs[i] = validationErr{
					err: errors.Wrap(err, "fail to write layer to workdir"),
				}
				return
			}

			tf, err := tfclient.New(tfWorkdir, tfpath)
			if err != nil {
				s.Error()
				errs[i] = validationErr{
					err: errors.Wrap(err, "fail to get terraform client"),
				}
				return
			}

			err = tf.Init(ctx, l.SHA)
			if err != nil {
				s.Error()
				errs[i] = validationErr{
					err: errors.Wrap(err, "fail to terraform init"),
				}
				return
			}

			validation, err := tf.Validate(ctx)
			if err != nil {
				s.Error()
				errs[i] = validationErr{
					err: errors.Wrapf(err, "fail to validate layer %s", l.Name),
				}
				return
			}

			if validation.ErrorCount > 0 {
				s.Error()

				var sb strings.Builder
				for _, d := range validation.Diagnostics {
					if d.Severity != tfjson.DiagnosticSeverityError {
						continue
					}

					fmt.Fprint(&sb, "\n  Error: ")
					fmt.Fprintf(&sb, "  %s\n\n", d.Summary)
					fmt.Fprintf(&sb, "    on %s line %d, in %s:\n", d.Range.Filename, d.Range.Start.Line, *d.Snippet.Context)
					fmt.Fprintf(&sb, "    %d: %s\n\n", d.Range.Start.Line, d.Snippet.Code)
				}

				errs[i] = validationErr{
					err:         errors.Errorf("layer %s is not valid", l.Name),
					diagnostics: sb.String(),
				}
				return
			}

			s.Complete()
		}(i, l)
	}

	wg.Wait()
	err = nil
	diagnostics := ""
	for i, e := range errs {
		if e.err != nil {
			diagnostics += fmt.Sprintf("\n%s diagnostics:\n%s", ls[i].Name, e.diagnostics)
			err = multierr.Append(err, e.err)
		}
	}

	if err != nil {
		sm.Stop()
		fmt.Fprint(os.Stdout, diagnostics)
		return err
	}

	savingSpinner := sm.AddSpinner("Saving layer definitions")

	location, err := c.definitionssBackend.Location(ctx)
	if err != nil {
		savingSpinner.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to get layers backend location")
	}

	err = c.definitionssBackend.UpdateLayers(ctx, ls)
	if err != nil {
		savingSpinner.Error()
		sm.Stop()
		return errors.Wrap(err, "fail to update layers")
	}

	savingSpinner.UpdateMessagef(
		"%d %s saved to \"%s\"",
		len(ls),
		pluralize("layer", len(ls)),
		location,
	)
	savingSpinner.Complete()
	sm.Stop()

	return nil
}

func pluralize(s string, n int) string {
	if n == 1 {
		return s
	}

	return s + "s"
}
