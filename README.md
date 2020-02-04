[![Build Status](https://travis-ci.org/instructure/muss.svg?branch=master)](https://travis-ci.org/instructure/muss)

# muss

Definition:

> noun: mess; scramble; a confused conflict
> verb: to put into disorder; to make messy

Or if you prefer, an acronym:

> Manage User Selected Services

`muss` is a command line application to allow users to choose what services
they want to run and how they want to run them.

`muss` is a wrapper around `docker-compose`.
When you run a `muss` subcommand it will:

- process (and validate) the config files
- generate a `docker-compose.yml`
- prepare any defined bind mounts (volumes)
- load secrets into the environment
- delegate to `docker-compose`

Projects (and their dependent services) can be configured
in a familiar style through a series of yaml files.


# Installation

    brew tap instructure/muss
    brew install muss


# Building, Testing, etc

`muss` is a CLI written in go.

`make install` to install it into your `$GOPATH`.

You can use `make test` to test all of the included go packages
or you can run any `go` commands directly (`go test ./cmd`).


# Usage

`muss help` will describe available subcommands.

Many docker-compose commands are supported directly
so that config files are always up to date and secrets
are loaded into the environment.

The simplest possible usage is to call `muss up`
to start all the services and type Ctrl-C to shut it down.

    $ muss help
    Configure and run project services

    Usage:
      muss [command]

    Available Commands:
      attach      Attach local stdio to a running container
      build       Build or rebuild services
      config      muss configuration
      dc          Call aribtrary docker-compose commands
      down        Stop and remove containers, networks, images, and volumes
      exec        Execute a command in a running container
      help        Help about any command
      logs        View output from services
      ps          List containers
      pull        Pull the latest images for services
      restart     Restart services
      rm          Remove stopped containers
      run         Run a one-off command
      start       Start services
      stop        Stop services
      up          Create and start containers
      version     Show version information
      wrap        Execute arbitrary commands

    Flags:
      -h, --help   help for muss

    Use "muss [command] --help" for more information about a command.

## Advanced Usage

For less common docker-compose commands that muss does not define
you can use the `dc` subcommand, e.g. `muss dc events --json`.

You can also run any arbitrary commands using `muss wrap`.
This allows you to run a command after files have been generated and
the environment has been loaded.

muss has its own `config` subcommand (different from the docker-compose
config command).

`muss config save` will generate the files.  This happens before running other
commands (like `muss up`) but this can be useful if you just want to inspect the
files.

`muss config show` will print out the whole configuration.  The `--format`
parameter takes a go template string to allow you to limit or manipulate the
config (useful for scripting and debugging).


# Configuration

A muss project is configured via a series of yaml files:

1. project config
2. user config
3. service definitions

The project configuration defines:

- default preferences for service configs
- secret commands
- the locations of additional config files

This file should be committed into source control.


A muss user file allows you to specify preferences and customizations:

- specify a different service preference order than the project's default
- disable services that you don't intend to use

This file should be ignored by version control.


The project file can include the location of service definition files,
each of which:

- has a name
- defines one or more config options for how to run the service
  (example use cases would be local repos, pre-baked docker images,
  or remote services).


## Project Config

The syntax and options for the main `muss.yaml` file:

```yaml
    ---
    # Path to user customization file.
    user_file: muss.user.yaml

    # If present will set COMPOSE_PROJECT_NAME (unless already set).
    project_name: myproject

    # Alternate target for compose config (default is docker-compose.yml).
    # If present will set COMPOSE_FILE (unless already set).
    # Note that when COMPOSE_FILE is set docker-compose will not automatically
    # use docker-compose.override.yml.
    # The override section of the muss user file will still work, however.
    compose_file: "docker-compose.muss.yml"

    # Define the order for which configuration option to use
    # for any service that has multiple options.
    default_service_preference:
      - registry
      - repo
      - internal

    # Service files are yaml files containing service definitions.
    service_files:
      - ./dev/database.yml
      - ./dev/microservice/service.yml

    # Secret commands define aliases that can be used by service definitions.
    secret_commands:
      vault:

        # Arguments will be prepended to the secret arguments.
        exec: ["vault", "kv", "get", "-field"]

        # Env Commands will be run once to setup the environment
        # if any secrets are requested.
        env_commands:
          # When you specify a varname the output of the command
          # will be set in that environment variable.  If that variable
          # is already set in the environment the command will not be run.
          - varname: VAULT_TOKEN
            exec: ["bin/vault-token"]

    # A passphrase is required for local caching of the secrets.
    # Use an env var representing your auth token.
    secret_passphrase: $VAULT_TOKEN

    # A status line will be fixed to the bottom of the screen during "up".
    status:
      # Stdout from this command will appear in the status line.
      exec: ["bin/muss-status"]
      # Each line of status output will be formatted with this spec:
      line_format: "# %s"
      # Specify how often to execute the command to update the status:
      interval: 5s
```


## User Config

Users can customize what they want to run with a user file:

```yaml
    ---
    # Users can set their own default order for which service config to prefer.
    # This list will be used before the default_service_preference of the
    # project config.
    service_preference:
      - remote
      - repo

    # Services can also be chosen specifically (to override preference lists)
    services:
      microservice:
        config: remote

      # ...or removed completely.
      stats:
        disabled: true

    # An override section can be defined that will be merged onto the
    # docker-compose config.  By defining it here, muss extensions (like file
    # volumes) can be utilized.
    override:
      services:
        app:
          environment:
            HOW_I_LIKE_IT: nifty
```


## Service Definitions

Service files contain service definitions that will generate chunks of a
docker-compose file (which will all be merged on top of each other).

A service can define multiple configs.
The names of the configs are arbitrary;
These are the values used in the
`service_preference` and `default_service_preference` lists.
Use common names among your services so that the
preference options apply as consistently as possible.

In the example below:
- "local" implies "run from local directory"
- "registry" implies "an image from a docker registry"
- "edge" and "staging" are different options for external integrations

When multiple configs are defined
the option will be chosen in this order:
- a specific user choice
- the first of any `service_preference`
- the first of any `default_service_preference`

The body of a service config can contain the following:
- "include" is a list of other config names to merge in before this one
  (merging always overwrites)
- "secrets" is a list of secrets to load
- "services" is a subset of the "services" section of a docker-compose
  configuration... it will be passed through.
- "volumes" is also just a piece of docker-compose syntax that will be passed.


```yaml
    ---
    # Service name.
    name: microservice

    configs:

      # Configs with a leading underscore are private/internal
      # and can be used as common bases for merging.
      _base:

        # The "services" map will be merged into the final docker-compose yaml.
        # We define attributes of multiple services here that will all be merged
        # on top of each other (the way a docker-compose override file works).
        services:
          app:
            environment:
              # The "app" service (defined more fully elsewhere)
              # will only recieve this environment variable
              # if this service definition is enabled.
              MICROSERVICE_ENABLED: '1'
          microservice:
            environment:
              APP_ENV: 'development'

      local:
        # A config can include other configs that will be merged in first.
        include:
          - _base

        # Any "volumes" will also pass to docker-compose.
        volumes:
          microservice-data: {}
        services:
          microservice:
            # Paths will be relative to the project root.
            build: ../microservice
            volumes:
              - microservice-data:/var/lib/microservice

      registry:
        include:
          - _base
        services:
          microservice:
            image: our-registry/microservice

      _remote:
        services:
          # This service can define how another service will reach it.
          app:
            environment:
              MICROSERVICE_URL:
              MICROSERVICE_KEY:
          # Note that there is no "microservice" here.
          # By pointing to a remote instance
          # we have eliminated the need to run it locally.

      edge:
        include:
          - _base
          - _remote
        # A config can define secrets that will be loaded into the environment.
        secrets:
          MICROSERVICE_URL: {vault: ["MICROSERVICE_URL", "app/edge/common"]}
          MICROSERVICE_KEY: {vault: ["MICROSERVICE_KEY", "app/edge/common"]}

      staging:
        include:
          - _base
          - _remote
        secrets:
          MICROSERVICE_URL: {vault: ["MICROSERVICE_URL", "app/staging/common"]}
          MICROSERVICE_KEY: {vault: ["MICROSERVICE_KEY", "app/staging/common"]}
```

Any "services" defined that do not contain a "build" or an "image" will not be
included in the final docker-compose file.  This allows one service to define
what secrets are available for another service even if that service won't be
used in the end.

Service integration can now be organized around how services will be selected.
All of the docker-compose configuration for a given service can be placed in the
same service file, including not only that service (or multiple services)
but also how it integrates with another service.

For example, if a microservice can be local, remote, or completely optional,
the definition file can point the environment variables for other services
to this one.  That way they update according to which ever option is chosen,
or the service can be completely disabled and the other services won't receive
the environment variables at all.

As another example, you could have a stats container that collects and prints
stats from all other services.  If the stats service file defines not only the
stats service but also the environment variables for all the other services
that will send stats to it, then you can simply disable the stats service
from your user config file and all the other services will no longer be
configured to send any stats.


# Secrets

Service definitions can specify secrets to be loaded by external commands.

The project configuration defines aliases that simplify the usage
and define commands that setup the environment
(for example, logging in to vault and returning the token).

When a service config is chosen that includes secrets:

- the setup commands are executed and the environment is populated
- the individual secret commands are then run and added as environment variables
- muss then delegates to the subcommand (docker-compose, etc).

Secrets are cached and encrypted with the `secret_passphrase`.
So if you use your auth token as your passphrase the secrets will be cached
for as long as your token is valid.  When you get a new token it will force
fetching new secrets.

Secret commands can either specify a `varname` and the STDOUT of the script
will be assigned to that var.
Alternatively the commands can specify: `parse: true`
and the output will be parsed as lines of `NAME=VALUE`.

STDIN and STDERR will pass directly so that users can response to password
prompts and see errors.

To provide a more concrete example:

`muss.yaml`:

    secret_commands:
      vault:
        exec: ["vault", "kv", "get", "-field"]
        env_commands:
          - exec: ["bin/vault-token"]
            parse: true
    # The passphrase will be parsed and env vars will be interpolated.
    secret_passphrase: $VAULT_TOKEN
    service_files:
      - dev/microservice.yml

The exec command can be something global
but will often be a command that is part of your project.

The `bin/vault-token` script would probably do something like this:

    #!/bin/bash

    # Setup any necessary env like (perhaps VAULT_ADDR).
    export VAULT_ADDR=...

    if ! token-is-valid; then
      vault login ...
      # persist the token and expiration so that next time
      # the script runs it won't have to login again.
    fi

    # These lines will be parsed and set in the environment.
    echo VAULT_ADDR="$VAULT_ADDR"
    echo VAULT_TOKEN="$(< ~/.vault-token )"

The actual secrets in the service definition (`dev/microservice.yml`):

    name: microservice
    configs:
      somewhere-far:
        secrets:
          SECRET_KEY: {vault: ["MICROSERVICE_KEY", "path/to/key"]}
        services:
          app:
            environment:
              SECRET_KEY: # no value, it will get it from the environment.

So when the "somewhere-far" option is chosen for the service
it will load those secrets.

First `bin/vault-token` will be executed and `VAULT_ADDR` and `VAULT_TOKEN`
will be set in the environment.

Then `vault kv get -field MICROSERVICE_KEY path/to/key` will be executed
and the value stored in the `SECRET_KEY` environment variable.

Then muss will continue and delegate to `docker-compose` to run your services
and the populated environment variables will be passed along.


# Additional Behavior

A few additional behaviors are defined beyond the normal docker-compose
features.


## Volumes

When bind mounts (host volumes) are specified muss will attempt to ensure
that they already exist.  This helps prevent permission problems that can occur
by letting the docker daemon create a directory under your project directory.


## File volumes

Additionally volumes can be specified to be files instead of directories
which can be useful for mounting config files into the container.
Just use the long syntax for volumes and add `file: true`:

    volumes:
      - target: /home/docker/.somerc
        source: ~/.somerc
        file: true

This way if the specified path isn't already a file muss will create it
so that docker doesn't make it a directory and cause confusing errors later.

The `file: true` will be removed from the resulting docker-compose file.
