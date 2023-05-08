# `preq` - command-line utility for all your pull request needs

> __Note__  
> ALPHA Please use with caution

`preq` is a command line utility for working with pull requests. It tries to be useful and quick by leveraging information from local Git repositories. Parameters can also be explicitly set using flags.

The application is in alpha state and only recommended for evaluation and only supports Bitbucket cloud currently.

## Installation

### MacOS using Homebrew

Homebrew tap
```bash
brew install tfrs1/tap/preq
```

### Linux

Linux builds can be found in [releases](https://github.com/tfrs1/preq/releases).

## Usage

`preq` is meant to be used in synthesis with Git. `preq` determines many parameters from the working directory (or parents) if it is also a Git repository. Of course, all parameters can be explicitly defined if needed.

For example, `preq` can find out the Git origin provider, the repository name, and the source branch for the `create` command.

The following global flags can be used with any `preq` command.
- `--provider`, `-p` - Provider, e.g. `bitbucket`
- `--repository`, `-r` - Repository name, e.g. `owner/repo-name`

### Terminal UI

`preq` is a TUI application as well as a CLI application. To start the TUI you can either run `preq` for the Git repository in the working directory or `preq -g` for all known Git repositories. `preq` keeps a history of all local repositories previously seen by `preq`.

> __Note__  
> Currently the only supported provider is Bitbucket cloud.
> - `bitbucket`

![TUI home](./docs/tui-home-screenshot.png)

### Commands

`preq` currently supports create, decline, approve, open, and list. Run `preq -h` to read more about them.

#### Default reviewers

Default reviewers will be automatically added to the pull requests created with `preq`. Since the program is not able to determine the UUID of your user, the PR creation request will fail if your user is one of the default reviewers. To fix this you need to add the UUID of your user to the configuration.

```toml
[bitbucket]
  username = "username"
  password = "user_password"
  uuid = "{universally-unique-identifier}"
```

Default reviewers will be automatically added to the pull requests created with `preq`.

## Configuration

It is possible to define the configuration in 3 formats, TOML, YAML and JSON. Global configuration file should be located in `~/.config/preq/config.toml` (or `config.yaml`, `config.json` for alternative formats). And per-repository configuration should be defined in `.preqcfg` file (any format) located in the root directory of a local Git repository (i.e. with the .git directory)

### Example config
```toml
[bitbucket]
  username = "bitbucket-username"
  password = "secret-password"
  aliases = [
    "bitbucket.org-work",
    "bitbucket.org-personal"
  ]
```
### Bitbucket cloud
To use Bitbucket cloud you must create an app password from the personal settings page with pull request read/write permissions.

* `aliases` - A list of hostname aliases for Bitbucket service. For example when using multiple accounts with different SSH keys.

## Roadmap

- [ ] Review pane improvements
  - [ ] Review progress
  - [ ] Mark the review state of individual files
  - [ ] Show changes since last review visit
- [ ] Add other providers
  - [ ] GitHub
  - [ ] GitLab
- [ ] Nerd font icons
- [ ] Intl support
- [ ] Keymap configuration
- [ ] Color configuration
- etc.

## Contributing

Contributions are welcome.

### Building from source

Clone the repository and run the build command.
```
go build
```
