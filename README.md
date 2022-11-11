# Sablier

![Github Actions](https://img.shields.io/github/workflow/status/acouvreur/sablier/Build?style=flat-square) ![Go Report](https://goreportcard.com/badge/github.com/acouvreur/sablier?style=flat-square) ![Go Version](https://img.shields.io/github/go-mod/go-version/acouvreur/sablier?style=flat-square) ![Latest Release](https://img.shields.io/github/release/acouvreur/sablier/all.svg?style=flat-square)

Sablier is an API that start containers for a given duration.

It provides an integrations with multiple reverse proxies and different loading strategies.

Which allows you to start your containers on demand and shut them down automatically as soon as there's no activity.

![Hourglass](https://raw.githubusercontent.com/acouvreur/sablier/main/docs/img/hourglass.png)

- [Sablier](#sablier)
  - [Quick start](#quick-start)
  - [Configuration](#configuration)
    - [Configuration File](#configuration-file)
    - [Environment Variables](#environment-variables)
    - [Arguments](#arguments)
  - [Install Sablier](#install-sablier)
    - [Use the Docker image](#use-the-docker-image)
    - [Use the binary distribution](#use-the-binary-distribution)
    - [Compile your binary from the sources](#compile-your-binary-from-the-sources)
  - [Loading with a waiting page](#loading-with-a-waiting-page)
    - [Dynamic Strategy Configuration](#dynamic-strategy-configuration)
    - [Creating your own loading theme](#creating-your-own-loading-theme)
  - [Blocking the loading until the session is ready](#blocking-the-loading-until-the-session-is-ready)
  - [Saving the state to a file](#saving-the-state-to-a-file)
  - [Reverse proxies integration plugins](#reverse-proxies-integration-plugins)
    - [Traefik Middleware Plugin](#traefik-middleware-plugin)
  - [Glossary](#glossary)
  - [Credits](#credits)

## Quick start

```bash
# Create and stop nginx container
docker run -d --name nginx nginx
docker stop nginx

# Create and stop whoami container
docker run -d --name whoami containous/whoami:v1.5.0
docker stop whoami

# Start Sablier with the docker provider
docker run -d -v /var/run/docker.sock:/var/run/docker.sock -p 10000:10000 ghcr.io/acouvreur/sablier:latest start --provider.name=docker

# Start the containers, the request will hang until both containers are up and running
curl 'http://localhost:10000/api/strategies/blocking?names=nginx&names=whoami&session_duration=1m'
{
  "session": {
    "instances": [
      {
        "instance": {
          "name": "nginx",
          "currentReplicas": 1,
          "desiredReplicas": 1,
          "status": "ready"
      },
        "error": null
      },
      {
        "instance": {
          "name": "nginx",
          "currentReplicas": 1,
          "desiredReplicas": 1,
          "status": "ready"
       },
        "error": null
      }
    ],
    "status":"ready"
  }
}
```

## Configuration

There are three different ways to define configuration options in Sablier:

1. In a configuration file
2. As environment variables
3. In the command-line arguments

These ways are evaluated in the order listed above.

If no value was provided for a given option, a default value applies.

### Configuration File

At startup, Sablier searches for configuration in a file named sablier.yml (or sablier.yaml) in:

- `/etc/sablier/`
- `$XDG_CONFIG_HOME/`
- `$HOME/.config/`
- `.` *(the working directory).*

You can override this using the configFile argument.

```bash
sablier --configFile=path/to/myconfigfile.yml
```

```yaml
provider:
  # Provider to use to manage containers (docker, swarm, kubernetes)
  name: docker 
server:
  # The server port to use
  port: 10000 
  # The base path for the API
  base-path: /
storage:
  # File path to save the state (default stateless)
  file:
sessions:
  # The default session duration (default 5m)
  default-duration: 5m
  # The expiration checking interval. 
  # Higher duration gives less stress on CPU. 
  # If you only use sessions of 1h, setting this to 5m is a good trade-off.
  expiration-interval: 20s
logging:
  level: trace
strategy:
  dynamic:
    # Custom themes folder, will load all .html files recursively (default empty)
    custom-themes-path:
    # Show instances details by default in waiting UI
    show-details-by-default: false
    # Default theme used for dynamic strategy (default "hacker-terminal")
    default-theme: hacker-terminal
    # Default refresh frequency in the HTML page for dynamic strategy
    default-refresh-frequency: 5s
  blocking:
    # Default timeout used for blocking strategy (default 1m)
    default-timeout: 1m
```

### Environment Variables

All environment variables can be used in the form of the config file such as 

```yaml
strategy:
  dynamic:
    custom-themes-path: /my/path
```

Becomes

```bash
STRATEGY_DYNAMIC_CUSTOM_THEMES_PATH=/my/path
```

### Arguments

To get the list of all available arguments:

```bash
sablier --help

# or

docker run acouvreur/sablier[:version] --help
# ex: docker run acouvreur/sablier:v1.0.0 --help
```

All arguments can be used in the form of the config file such as 

```yaml
strategy:
  dynamic:
    custom-themes-path: /my/path
```

Becomes

```bash
sablier start --strategy.dynamic.custom-themes-path /my/path
```

## Install Sablier

You can install Traefik with the following flavors:

- Use the Docker image
- Use the binary distribution
- Compile your binary from the sources

### Use the Docker image

- **Docker Hub**: [acouvreur/sablier](https://hub.docker.com/r/acouvreur/sablier)
- **Github Container Registry**: [ghcr.io/acouvreur/sablier](https://github.com/acouvreur/sablier/pkgs/container/sablier)
  
Choose one of the Docker images and run it with one sample configuration file:

- [sablier.yaml](https://raw.githubusercontent.com/acouvreur/sablier/main/sablier.sample.yaml)

```bash
docker run -d -p 10000:10000 \
    -v $PWD/sablier.yml:/etc/sablier/sablier.yml acouvreur/sablier:1.1.0
```

### Use the binary distribution

Grab the latest binary from the [releases](https://github.com/acouvreur/sablier/releases) page.

And run it:

```bash
./sablier --help
```

### Compile your binary from the sources

```bash
git clone git@github.com:acouvreur/sablier.git
cd sablier
make
# Output will change depending on your distro
./sablier_draft_linux-amd64
```

## Loading with a waiting page

**The Dynamic Strategy provides a waiting UI with multiple themes.**
This is best suited when this interaction is made through a browser.

|       Name        |                                                  Preview                                                  |
| :---------------: | :-------------------------------------------------------------------------------------------------------: |
|      `ghost`      |           ![ghost](https://raw.githubusercontent.com/acouvreur/sablier/main/docs/img/ghost.png)           |
|     `shuffle`     |         ![shuffle](https://raw.githubusercontent.com/acouvreur/sablier/main/docs/img/shuffle.png)         |
| `hacker-terminal` | ![hacker-terminal](https://raw.githubusercontent.com/acouvreur/sablier/main/docs/img/hacker-terminal.png) |
|     `matrix`      |          ![matrix](https://raw.githubusercontent.com/acouvreur/sablier/main/docs/img/matrix.png)          |

### Dynamic Strategy Configuration

```yaml
strategy:
  dynamic:
    # Custom themes folder, will load all .html files recursively (default empty)
    custom-themes-path:
    # Show instances details by default in waiting UI
    show-details-by-default: false
    # Default theme used for dynamic strategy (default "hacker-terminal")
    default-theme: hacker-terminal
    # Default refresh frequency in the HTML page for dynamic strategy
    default-refresh-frequency: 5s
```

### Creating your own loading theme

See [creating your own theme](docs/THEMES.md).

## Blocking the loading until the session is ready

**The Blocking Strategy waits for the instances to load before serving the request**
This is best suited when this interaction from an API.

```yaml
strategy:
  blocking:
    # Default timeout used for blocking strategy (default 1m)
    default-timeout: 1m
```

## Saving the state to a file

You can save the state of the application in case of failure to resume your sessions.

For this you can use the `storage` configuration.

```yml
storage:
  file: /path/to/file.json
```

If the file doesn't exist it will be created, and it will be syned upon exit.

Loaded instances that expired during the restart won't be changed though, they will simply be ignored.

## Reverse proxies integration plugins

### Traefik Middleware Plugin

See [Traefik Middleware Plugin](https://github.com/acouvreur/sablier/tree/main/plugins/traefik/README.md)

## Glossary

I'll use these terms in order to be provider agnostic.

- **Session**: A Session is a set of **instances**
- **Instance**: An instance is either a docker container, docker swarm service, kubernetes deployment or kubernetes statefulset

## Credits

- [Hourglass icons created by Vectors Market - Flaticon](https://www.flaticon.com/free-icons/hourglass)
- [tarampampam/error-pages](https://github.com/tarampampam/error-pages/) for the themes