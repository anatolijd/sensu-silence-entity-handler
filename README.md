[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/anatolijd/sensu-silence-entity-handler)
![Go Test](https://github.com/anatolijd/sensu-silence-entity-handler/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/anatolijd/sensu-silence-entity-handler/workflows/goreleaser/badge.svg)

# Sensu Silence Event Handler

## Table of Contents
- [Overview](#overview)
- [Usage](#usage)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
- [Installation from source](#installation-from-source)
- [Contributing](#contributing)

## Overview

The Sensu Silence Event Handler is a simple handler that selinces entities from the
Sensu Entities API. Any valid Sensu Event can be used to initiate entity silence,
including keepalive events.

## Usage

```
Silence Sensu entities on-demand. It does not perform any validation. It simply consumes events and silences entity referenced in the event.

Usage:
 sensu-silence-entity-handler [flags]
 sensu-silence-entity-handler [command]

Available Commands:
 help        Help about any command
 version     Print the version number of this plugin

Flags:
     --access-token string        Sensu Access Token
     --api-key string             Sensu API Key
     --api-url string             Sensu API URL (default "http://127.0.0.1:8080")
 -e, --expire int                 silence period, seconds (default 1800)
 -h, --help                       help for sensu-silence-entity-handler
 -i, --insecure-skip-tls-verify   skip TLS certificate verification
     --namespace string           Sensu Namespace
 -r, --reason string              Reason (default "sensu-silence-entity-handler")
     --trusted-ca-file string     Sensu Trusted Certificate Authority file

Use "sensu-silence-entity-handler [command] --help" for more information about a command.
```

## Configuration

### Asset registration

[Sensu Assets][10] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add anatolijd/sensu-silence-entity-handler
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][https://bonsai.sensu.io/assets/anatolijd/sensu-silence-entity-handler].

### Handler definition

```yml
---
type: Handler
api_version: core/v2
metadata:
  name: sensu-silence-entity-handler
spec:
  type: pipe
  command: sensu-silence-entity-handler
  runtime_assets:
  - anatolijd/sensu-silence-entity-handler
```

## Installation from source

The preferred way of installing and deploying this plugin is to use it as an Asset. If you would
like to compile and install the plugin from source or contribute to it, download the latest version
or create an executable script from this source.

From the local path of the sensu-silence-entity-handler repository:

```
go build
```

## Contributing

For more information about contributing to this plugin, see [Contributing][1].

[1]: https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md
[2]: https://github.com/sensu-community/sensu-plugin-sdk
[3]: https://github.com/sensu-plugins/community/blob/master/PLUGIN_STYLEGUIDE.md
[4]: https://github.com/sensu-community/handler-plugin-template/blob/master/.github/workflows/release.yml
[5]: https://github.com/sensu-community/handler-plugin-template/actions
[6]: https://docs.sensu.io/sensu-go/latest/reference/handlers/
[7]: https://github.com/sensu-community/handler-plugin-template/blob/master/main.go
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-community/sensu-plugin-tool
[10]: https://docs.sensu.io/sensu-go/latest/reference/assets/
