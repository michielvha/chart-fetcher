# Changelog

## Version 0.0.2
### ⚙️ Patches
- **Code cleanup**
  - Combined `AddRepo` & `FetchRepoIndex` into a single method.
  - Decoupled `Login` & `AddRepo` methods to allow for proper authentication. OCI uses `login + OCI pull`, while legacy uses `AddRepo + legacy pull`.
  - Made the code more modular by creating helper functions for common tasks.
  - seperated helm functions into separate files for readability.
  - removed `ioutil` package because [deprecated & redundant](https://go.dev/doc/go1.16#ioutil), incorporated into and since maintained in `io` & `os`.
## Version 0.0.1
### 🚀 Major Release
- **Initial Release of ChartFetch**
  - Added support for json & yaml configuration files
  - Added support for environment variables
  - Added support for command line arguments
    - `--config`: Specify a custom configuration file path. Defaults to `config.json`
    - `--outputPath`: Specify a custom output path for the charts. Default to `charts/`
  - Added support for OCI registries through 
    - [`Login`](): To conform to the new OCI standard we must use something similar to helm login. `h.RegistryClient.Login` method is used to login to the registry.
    - [`PullOCIChart`](): Pulling charts via the SDK using the `h.RegistryClient.Pull` method.
  - Added support for Legacy repositories through: 
    - [`AddRepo`](): Legacy repositories require you to add the repo to the helm client.
    - [`FetchRepoIndex`](): Repo index file needs to be fetched by the binary to be able to search the repo.
    - [`EnsureRepoFileExists`](): check if the initial `repostories.yaml` file exists, if not create it.
    - [`PullLegacyChart`](): use the old method via http to pull the chart by reading the index file.
