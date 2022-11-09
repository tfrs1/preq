# `preq` - command-line utility for all your pull request needs

`preq` tries to be useful and smart about creating your pull requests. It attempts to determine all parameters from Git repository in the working directory where the command is executed. All parameters can be, of course, overridden using flags.

## Installation

### MacOS using Homebrew

Homebrew tap
```bash
brew install tfrs1/tap/preq
```

### Linux

Linux builds can be found in [releases](https://github.com/tfrs1/preq/releases).

## Usage

`preq` is meant to be used in synthesis with Git. `preq` determines many parameters from the working directory if it is also a Git repository. Of course, all parameters can be explicitly defined if needed.

For example, `preq` can find out the Git origin provider, the repository name, and the source branch for the `create` command.

Most of the commands support support the following optional so they will be omitted from command specific documentation.

- `--provider`, `-p` - Provider, e.g. `bitbucket`
- `--repository`, `-r` - Repository name, e.g. `owner/repo-name`

> __Note__  
> Currently the only supported provider is Bitbucket cloud.
> - `bitbucket`

### Bitbucket password

Create app password in Bitbucket with pull request read/write permissions. User ID permissions unless the ID is added to the configuration.

### Create

The create command supports the following flags, but none of them are required. The create command is in some cases able to determine all parameters based on the local Git repository.

- `--destination`, `-d` - Destination branch name
- `--source`, `-s` - Source branch name
- `--draft` - Marks the pull request as work in progress

#### Default reviewers

Default reviewers will be automatically added to the pull requests created with `preq`.

#### Git repository example
```bash
preq create -d master
```
In the future the `destination` flag will also be optional, and it will default to either `master` or `develop` depending on which is the closest parent. This will also be configurable per Git repository.

#### Full command example
```bash
preq create -p bitbucket -r owner/repo -s develop -d master
```

### Open

Opens the pull request page.

Flags:
- `--print` - Prints out the web page URL instead of opening it

Example
```bash
preq open
```

## Configuration

toml in `.preqcfg` for per dir config or `~/.config/preq/config.toml` for global.

### Example config
```toml
[bitbucket]
  username = "bitbucket-username"
  password = "secret-password"

[templates]
  draft = "%s - Draft"
```

Reviewers?
Close branch?
Merge strategy?

## Future additions

- Add other commands (decline, accept, merge, info, etc.)
- Add other providers (GitHub, GitLab, etc.)
- Add interactive mode (--interactive)
- Docker image

## Contributing

Contributions are welcome.

### Building from source

Clone the repository and run the build command.
```
go build
```
