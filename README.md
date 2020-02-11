# `prctl` - command-line utility for all your pull request needs

`prctl` tries to be useful and smart about creating your pull requests. It attempts to determine all parameters from Git repository in the working directory where the command is executed. All parameters can be, of course, overridden using flags.

## Installation

### MacOS using Homebrew

Homebrew tap
```bash
brew install tfrs1/tfrs1/prctl
```

### Linux

Linux builds can be found in [releases](https://github.com/tfrs1/prctl/releases).

## Usage

`prctl` is meant to be used in synthesis with Git. `prctl` determines many parameters from the working directory if it is also a Git repository. Of course, all parameters can be explicitly defined if needed.

For example, `prctl` can find out the Git origin provider, the repository name, and the source branch for the `create` command.

Currently the only supported provider is Bitbucket cloud.
- `bitbucket-cloud`

### Create

The create command supports the following flags, but none of them are required. The create command is in some cases able to determine all parameters based on the local Git repository.

- `--destination`, `-d` - Destination branch name
- `--source`, `-s` - Source branch name
- `--provider`, `-p` - Provider, e.g. `bitbucket-cloud`
- `--repository`, `-r` - Repository name, e.g. `owner/repo-name`

#### Git repository example
```bash
prtcl create -d master
```
In the future the `destination` flag will also be optional, and it will default to either `master` or `develop` depending on which is the closest parent. This will also be configurable per Git repository.

#### Full command example
```bash
prtcl create -p bitbucket-cloud -r owner/repo -s develop -d master
```

## Future additions

- Add other commands (decline, accept, merge, info, list, etc.)
- Add other providers (GitHub, GitLab, etc.)
- Add interactive mode (--interactive)


## Contributing

### Building from source

Clone the repository and run the build command.
```
go build
```
