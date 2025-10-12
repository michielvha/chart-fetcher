# ChartFetch

The `chartfetch` Go application is a solution designed to address a limitation in our GitOps workflow, which utilizes **ArgoCD** and **Kustomize**. 

Our challenge lies in pulling Helm charts from an authenticated registry (Jfrog), as ArgoCD's Kustomize plugin does not support chart retrieval from registries requiring authentication. To bridge this gap, `chartfetch` acts as an intermediary application capable of:
- Authenticating with both legacy and OCI-compliant Helm registries.
- Pulling Helm charts and making them accessible for seamless integration into the GitOps workflow.

This application is designed to run alongside a web server, serving as a reliable and scalable solution to facilitate Helm chart management in complex, authenticated environments.

### Example Run Scenarios

| **Scenario**                      | **Result**                                     |
|-----------------------------------|------------------------------------------------|
| No input                          | Uses `./config.json` and `charts/` as default. |
| `--config /path/to/config.json`   | Uses `/path/to/config.json`.                   |
| `CONFIG_PATH=/path/from/env.json` | Uses `/path/from/env.json`.                    |
| `--outputPath /path/to/output`    | Uses `/path/to/output`.                        |
| `OUTPUT_PATH= /path/to/output`    | Uses `/path/from/env.json`.                    |
| Both `--parameter` and `ENV_VAR`  | Environment variable takes precedence.         |


### Example Configuration File

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
  - url: "quay.io/kannika/charts"
    username_env: "REG1_USERNAME"
    password_env: "REG1_PASSWORD"
    charts:
      - name: "kannika"
        version: "0.9.1"
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
helm registry login myreg.azurecr.io --username $REG1_USERNAME --password-stdin $REG1_PASSWORD
````

#### initialize Go project using GitHub

```golang
go mod init github.com/michielvha/ChartFetch
```

# upcoming features
- **Pipeline:**
  - linting
  - vulnerability scan
  - auto gen release notes
  - [] Generate container
- **Application:**
  - proper output on run, add customization
  - [x] create container template
- **Docs:**
  - pipeline docs
  - application docs

