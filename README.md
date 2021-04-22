# tf-init-booster

When running terraform on a project with many modules, sourced from git, the operation to run `terraform init` becomes very slow. This tool will speedup that process by cloning the git repositoy once, and copy a checked out working tree to the `.terraform/modules` folder. Any module references pointing to the same repository and branch are using a symlink instead of duplicating the module sources.

## Usage

Run `tf-init-booster` before you run `terraform init`. The tool will download all required modules in advance.

## Git-Crypt

When the environment variable `GIT_CRYPT_KEY` is set, the tool will execute `git-crypt unlock` after downloading the module sources.

## License

This software is released under the Apache 2.0 License.
