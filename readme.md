# **Charthost**

The `Charthost` Go application is a solution designed to address a limitation in our GitOps workflow, which utilizes **ArgoCD** and **Kustomize**. 

Our challenge lies in pulling Helm charts from an authenticated registry (Jfrog), as ArgoCD's Kustomize plugin does not support chart retrieval from registries requiring authentication. To bridge this gap, `Charthost` acts as an intermediary application capable of:
- Authenticating with both legacy and OCI-compliant Helm registries.
- Pulling Helm charts and making them accessible for seamless integration into the GitOps workflow.

This application is designed to run alongside a web server, serving as a reliable and scalable solution to facilitate Helm chart management in complex, authenticated environments.

[[_TOC_]]

### **Example Run Scenarios**

| **Scenario**                      | **Result**                                     |
|-----------------------------------|------------------------------------------------|
| No input                          | Uses `./config.json` and `charts/` as default. |
| `--config /path/to/config.json`   | Uses `/path/to/config.json`.                   |
| `CONFIG_PATH=/path/from/env.json` | Uses `/path/from/env.json`.                    |
| `--outputPath /path/to/output`    | Uses `/path/to/output`.                        |
| `OUTPUT_PATH= /path/to/output`    | Uses `/path/from/env.json`.                    |
| Both `--parameter` and `ENV_VAR`  | Environment variable takes precedence.         |


### **Example Configuration File**

Configuration files can be supplied using either ``json`` or ``yaml`` formats.

#### ``JSON`` example:

````json
{
  "registries": [
    {
      "url": "quay.io/kannika/charts",
      "is_oci": true,
      "charts": [
        { "name": "kannika", "version": "0.9.1" }
      ]
    },
    {
      "url": "https://engiebnlms.jfrog.io/artifactory/api/helm/prd-helm-virtual",
      "username_env": "REG1_USERNAME",
      "password_env": "REG1_PASSWORD",
      "charts": [
        { "name": "autopass", "version": "0.1.0" }
      ]
    },
    {
      "url": "https://engiebnlms.jfrog.io/artifactory/api/helm/prd-helm-external-secrets",
      "username_env": "REG1_USERNAME",
      "password_env": "REG1_PASSWORD",
      "charts": [
        { "name": "external-secrets", "version": "0.10.7" }
      ]
    }
  ]
}
````

#### ``YAML`` example:

````yaml
registries:
  - url: "quay.io/kannika/charts"
    is_oci: true
    charts:
      - name: "kannika"
        version: "0.9.1"
  - url: "https://engiebnlms.jfrog.io/artifactory/api/helm/prd-helm-virtual"
    username_env: "REG1_USERNAME"
    password_env: "REG1_PASSWORD"
    charts:
      - name: "autopass"
        version: "0.1.0"
  - url: "https://engiebnlms.jfrog.io/artifactory/api/helm/prd-helm-external-secrets"
    username_env: "REG1_USERNAME"
    password_env: "REG1_PASSWORD"
    charts:
      - name: "external-secrets"
        version: "0.10.7"
````

### Troubleshooting

#### Gather official docs

````shell
go doc "packagename"

# example:
go doc helm.sh/helm/v3/pkg/repo
````

#### login with helm to ACR

````shell
helm registry login mslocalfoundationacr.azurecr.io --username $REG1_USERNAME --password-stdin $REG1_PASSWORD
````

#### Login to Artifactory

You seem to need to add repo to be able to pull from Artifactory

````shell 

````

#### initialize Go project using Azure-Devops

```golang
go mod init dev.azure.com/bnl-ms/AzureFoundation/charthost
```

# upcoming features
- **Pipeline**:
  - linting
  - vulnerability scan
  - auto gen release notes
  - Generate container
- **Application**:
  - proper output on run
  - create container template

# Changelog

## Version 1.0.1
### ⚙️ Patches
- **Code cleanup**
  - Combined `AddRepo` & `FetchRepoIndex` into a single method.
  - Decoupled `Login` & `AddRepo` methods to allow for proper authentication. OCI uses `login + OCI pull`, while legacy uses `AddRepo + legacy pull`.
  - Made the code more modular by creating helper functions for common tasks.
## Version 1.0.0
### 🚀 Major Release
- **Initial Release of Charthost**
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
