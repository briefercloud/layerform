package kill

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"

	"github.com/ergomake/layerform/pkg/layerdefinitions"
	"github.com/ergomake/layerform/pkg/layerinstances"
)

type DependenceInfo struct {
	DefinitionName string
	InstanceName   string
}

func HasDependants(
	ctx context.Context,
	instancesBackend layerinstances.Backend,
	definitionsBackend layerdefinitions.Backend,
	layerName, instanceName string,
) (bool, error) {
	hclog.FromContext(ctx).Debug("Checking if layer has dependants", "layer", layerName, "instance", instanceName)

	definitions, err := definitionsBackend.ListLayers(ctx)
	if err != nil {
		return false, errors.Wrap(err, "fail to list layers")
	}

	for _, definition := range definitions {
		isChild := false
		for _, d := range definition.Dependencies {
			if d == layerName {
				isChild = true
				break
			}
		}

		if isChild {
			instances, err := instancesBackend.ListInstancesByLayer(ctx, definition.Name)
			if err != nil {
				return false, errors.Wrap(err, "fail to list layer instances")
			}

			for _, instance := range instances {
				parentInstanceName := instance.GetDependencyInstanceName(layerName)
				if parentInstanceName == instanceName {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func GetDependants(
	ctx context.Context,
	instancesBackend layerinstances.Backend,
	definitionsBackend layerdefinitions.Backend,
	layerName, instanceName string,
	visited map[string]bool,
) ([]DependenceInfo, error) {
	hclog.FromContext(ctx).Debug("Finding dependant layers", "layer", layerName, "instance", instanceName)

	definitions, err := definitionsBackend.ListLayers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "fail to list layers")
	}

	dependantLayers := []DependenceInfo{}

	visited[layerName] = true

	for _, definition := range definitions {
		isChild := false
		for _, d := range definition.Dependencies {
			if d == layerName {
				isChild = true
				break
			}
		}

		if isChild {
			instances, err := instancesBackend.ListInstancesByLayer(ctx, definition.Name)
			if err != nil {
				return nil, errors.Wrap(err, "fail to list layer instances")
			}

			for _, instance := range instances {
				parentInstanceName := instance.GetDependencyInstanceName(layerName)
				if parentInstanceName == instanceName {
					dependantLayers = append(dependantLayers, DependenceInfo{
						DefinitionName: definition.Name,
						InstanceName:   instance.InstanceName,
					})
				} else if !visited[definition.Name] {
					childDependantLayers, err := GetDependants(
						ctx,
						instancesBackend,
						definitionsBackend,
						definition.Name,
						instance.InstanceName,
						visited,
					)
					if err != nil {
						return nil, err
					}
					dependantLayers = append(dependantLayers, childDependantLayers...)
				}
			}
		}
	}

	return dependantLayers, nil
}
