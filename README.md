# Kinde Terraform Provider

See [ORIGINAL.md](./ORIGINAL.md) for terraform plugin framework documentation.

## Prerequisites

Ensure you have credentials per [the prerequisites](https://github.com/axatol/kinde-go#prerequisites)

## quickstart

The minimum configuration

```terraform
terraform {
  required_providers {
    kinde = {
      source = "axatol/kinde"
    }
  }
}

# load everything from env
provider "kinde" {
}

# customise fields
provider "kinde" {
  alias    = development
  domain   = "https://example-development.au.kinde.com"
  audience = "https://example-development.au.kinde.com/api"
}
```
