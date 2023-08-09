<p align="center">
  <a href="https://layerform.dev">
  <picture>
    <source width="500px" media="(prefers-color-scheme: dark)" srcset="./assets/img/logo-square.png">
    <source width="500px" media="(prefers-color-scheme: light)" srcset="./assets/img/logo-square.png">
    <img width="500px" alt="layerform logo" src="./assets/img/logo-square.png">
    </picture>
  </a>
</p>

<h4 align="center">
  <a href="https://layerform.dev">Home Page</a> |
  <a href="https://docs.layerform.dev">Documentation</a> |
  <a href="https://discord.gg/daGzchUGDt">Discord</a> |
  <a href="https://docs.layerform.dev">Blog</a>
</h4>

<h1 align="center">
  Layerform
</h1>

<p align="center">
    <strong>
        Layerform helps engineers create their own staging environment using plain Terraform files.
    </strong>
</p>

<p align="center">
  <a href="https://github.com/ergomake/layerform/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/ergomake/layerform" alt="Layerform is released under the GNU GPLv3 license." />
  </a>
  <a href="https://discord.gg/daGzchUGDt">
    <img src="https://img.shields.io/discord/1055836143733706874" alt="Discord Chat" />
  </a>
  <a href="https://twitter.com/intent/follow?screen_name=GetErgomake">
    <img src="https://img.shields.io/twitter/follow/GetErgomake.svg?label=Follow%20@GetErgomake" alt="Follow @GetErgomake" />
  </a>
</p>

## What is Layerform?

Layerform makes it easy for engineers to create reusable layers of infrastructure using plain Terraform files.

When using Layerform, engineers encapsulate each part of their infrastructure into layer definitions.

Engineers can create infrastructure by stacking each of those layers. Layerform's magic is that layers can share the same base layers, allowing for easy and inexpensive reuse.

<p align="center">
  <img width="600px" src="./assets/img/all-layers.png" />
</p>

## Use cases

### Development environments

Engineers can use Layerform to quickly spin up production-like environments.

Development environments created with Layerform are similar to production environments because they use the same Terraform files.

Additionally, Layerform's development enviroments are less costly and quicker to spin up because they can reuse the same base layers of infrastructure. With Layerform, engineers only apply the infrastructure layers they need.

<p align="center">
  <img width="600px" src="./assets/img/dev-environments.png" />
</p>

### Encapsulation / Isolation of concerns

Another advantage of breaking infrastructure into layers is that organizations can define clearer boundaries between teams. Consequently, it will be easier for these organizations to [mirror their team's structure into their system's structure](https://martinfowler.com/bliki/ConwaysLaw.html).

<p align="center">
  <img width="600px" src="./assets/img/layers-vs-org.png" />
</p>

### Cost attribution and cost savings

In addition to saving costs by reusing infrastructure, Layerform allows you to automatically track costs for each layer instance.

When applying layers, Layerform will automatically tag the resources it creates with the actual name assigned to the layer instance. If you have `production` and `development` base layers, for example, each of those two will contain the tag `layerform_name` with their respective names.

Then, each resource on top of those base layers will include `layerform_base_name` with its respective base layer. For example, if multiple developers are spinning up resources on top of the `development` base layer, their own resources will contain a `layerform_base_name` tag whose value is `development`.

That way, Layerform can recursively traverse layers' resources to collect cost management information. Consequently, it will be able to tell the cost of your whole `production` and `development` layers, as well as an aggregate cost report of everything on top of those layers.

## Getting started

The first step to use Layerform is to create the Terraform files to provision each layer of infrastructure.

```
terraform/
├── layers/
│   ├── base_layer/
│   │   ├── eks.tf
│   │   └── kafka.tf
│   ├── services_layer/
│   │   ├── topic.tf
│   │   └── services.tf
│   └── integrations_layer/
│       ├── s3.tf
│       ├── lambda.tf
│       ├── topic.tf
│       └── services.tf
└── main.tf
```

Once you have your infrastructure defined as code, you'll use Terraform and the `layerform-provider` to determine each layer's name and files.

```hcl
# In main.tf

terraform {
  required_providers {
    layerform = {
      source  = "ergomake/layerform"
      version = "~> 0.1"
    }
  }
}

provider "layerform" {
  backend = "local"
}

resource "layerform_layer" "base_layer" {
  name   = "base"
  files = ["./base_layer/**"]
}

resource "layerform_layer" "services_layer" {
  name   = "services"
  files = ["./services_layer/**"]
  dependencies = [
    layerform_layer.base_layer.id
  ]
}

resource "layerform_layer" "integrations_layer" {
  name   = "integrations"
  files = ["./integrations_layer/**"]
  dependencies = [
    layerform_layer.base_layer.id
  ]
}
```

After defining each layer, you should `terraform apply` them. The `layerform-provider` will then take care of creating unique IDs for each layer and sending the Terraform files' contents to the Layerform Back-end.

After saving these layer definitions, you can use `layerform spawn <layer_name>` to create an instance of that particular layer. Each instance of a layer contains all the pieces of infrastructure defined within that layer's files.

For example, to create infrastructure for the `services` layer you should run `layerform spawn services`. That command will cause `layerform` to:

1. Create infrastructure for the `base` layer and assign it the ID `default`.
2. Create the infrastruture for the `services` layer on top of the `default` instance of the `base` layer
3. Assign a random ID to your `services` layer.

<p align="center">
  <img width="350px" src="./assets/img/default-base-layer.png" />
</p>

To spawn yet another `services` layer, just run `layerform spawn services` layer again. By default, Layerform will try to use underlying layers whose ID is `default` as base layers. Again, your `services` layer instance will be assigned a random ID.

<p align="center">
  <img width="350px" src="./assets/img/multiple-top-layers.png" />
</p>

> As a general rule, underlying layers are always the ones whose ID is `default`. Only the target layer gets a random ID.

To specify the desired ID for each underlying layer, you'll have to use the `--id-[layername] <id>`. For example:

```
# Creates:
# 1. A base layer with ID "one"
# 2. A services layer with ID "two"

$ layerform spawn services two --id-base=one
```

<p align="center">
  <img width="350px" src="./assets/img/one-two-layers.png" />
</p>

## Reusing infrastructure for development environments

Let's assume you wanted to create a whole separate environment for engineers to develop their applications against. This environment needs to closely resemble production, but it can't interfere with production resources like the production Kubernetes and Kafka instances.

For that, you can use `layerform spawn base development`. This command will cause `layerform` to create brand new resources corresponding to the `base` layer, and group them under the ID `development`.

<p align="center">
  <img width="450px" src="./assets/img/dev-single-layers.png" />
</p>

After that, multiple back-end engineers can spin-up their development infrastructure on top of this layer's resources. For example, an engineer called Alice could use `layerform spawn services alice-dev --x-base=development` to create their infrastructure, while an engineer called Bob could use `layerform spawn services bob-dev --x-base=development`.

<p align="center">
  <img width="600px" src="./assets/img/dev-multi-layers.png" />
</p>

## Layer immutability and layer rebasing

A layer can only mutate itself or the layers above. For example, if you have a `base` layer and a `backend` layer, the `backend` layer's Terraform files will _not_ be able to mutate any infrastructure in a `base` layer instance. Still, the `base` layer files can mutate any instances of the layers above it.

The way Layerform prevents undesirable mutations is by analyzing each `terraform plan` and detecting whether any mutation's target belongs to an underlying layer.

The reason Layerform prevents a layer from mutating its underlying layer is to avoid breaking sibling pieces of infrastructure.

This design allows for platform teams to "rebase" layer instances on top of theirs. For example, assume you have multiple application layers on top of a Kubernetes cluster belonging to a `base` layer. In that case, if the platform team wants to update the Kubernetes version and needs to patch existing application's manifests, they can do so from their own layer by referencing and patching other Terraform resources.

On the other hand, product engineers on the layers above cannot modify the `base` layer containing the Kubernetes cluster. Otherwise, they could break everyone else's applications.

In addition to preventing failures, immutability defines clearer communication interfaces between teams and helps organizations avoid lots of lateral channels.

## How Layerform works

Layerform has three major components. The `layerform-provider`, the Layerform Back-end, and Layerform CLI.

<p align="center">
  <img width="700px" src="./assets/img/all-components.png" />
</p>

The `layerform-provider` is used by Terraform to provision the Layerform Back-end with all the metadata for each layer, like its name and dependencies, and all the Terraform files associated with that layer.

The Layerform Back-end stores the data for each layer definition and stores the state for each instance of each layer so that new layers know which base state to use.

> There can be multiple types of back-ends. The most common types of back-end are `local`, for storing data locally, and `ergomake`, for storing data on the cloud.

Finally, the Layerform CLI talks to the Layerform Back-end to fetch the files for the layer it wants to apply, and the state for the underlying layer.

The way the Layerform CLI creates new layers on top of the correct existing layers is by injecting the underlying layer's state when applying each layer.

## Layerform design philosophy

Our main goal with Layerform was to make it as easy as possible for engineers to create and share different parts of their infrastructure. That way, we'd empower teams to create their own environments without burdening their organization with unnecessary costs or complex configuration files.

When developing Layerform, we also determined it should support virtually _any_ type of infrastructure, including infrastructure for serverless applications. That's why we decided to create a wrapper on top of Terraform, which supports Kubernetes/Helm, and already has established providers for all major public clouds.

Third, we decided Layerform should be simple and intuitive. Engineers shouldn't have to learn new proprietary languages or configuration formats to use Layerform. Whenever possible, we should allow them to reuse their existing configurations. Layerform concepts are the only thing engineers will need to learn about. Everything else should be "just Terraform".

Finally, we decided Layerform needs to be open and free. It's for that reason we're using a GPL license, and that's why you don't necessarily need to pay for anything before you can extract value from Layerform. Sure, Layerform Cloud can make things easier and provide a bunch of interesting Governance and Management features, but those are not necessary.

## Issues & Support

You can find Layerform's users and maintainers in [GitHub Discussions](https://github.com/ergomake/layerform/discussions). There you can ask how to set up Layerform, ask us about the roadmap, and discuss any other related topics.

You can also reach us directly (and more quickly) on our [Discord server](https://discord.gg/daGzchUGDt).

## Other channels

-   [Issue Tracker](https://github.com/ergomake/layerform/issues)
-   [Twitter](https://twitter.com/GetErgomake)
-   [LinkedIn](https://www.linkedin.com/company/layerform)
-   [Ergomake Engineering Blog](https://ergomake.dev/blog)

## License

Licensed under the [GNU GPLv3 License](https://github.com/layerform/layerform/blob/main/LICENSE).
