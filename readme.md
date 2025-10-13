# Chart Fetcher


[![Build and Release](https://github.com/michielvha/chart-fetcher/actions/workflows/build-release.yml/badge.svg)](https://github.com/michielvha/chart-fetcher/actions/workflows/build-release.yml)
[![Release](https://img.shields.io/github/release/michielvha/chart-fetcher.svg?style=flat-square)](https://github.com/michielvha/chart-fetcher/releases/latest)
[![Go Report Card]][go-report-card]

The `chart-fetcher` Go application is a solution designed to address a limitation in **Kustomize** based GitOps workflows. 

The challenge lies in pulling Helm charts from an authenticated registry, as Kustomize's helm plugin does not support chart retrieval from registries requiring authentication. To bridge this gap, `chart-fetcher` acts as an intermediary application capable of:
- Authenticating with both legacy and OCI-compliant Helm registries.
- Pulling Helm charts and making them accessible.

This application is designed to run alongside a web server, serving as a reliable and scalable solution to facilitate Helm chart management in authenticated environments.

> [!NOTE]
> This isn't trying to be a full-featured Helm client. It does one thing well: pulling charts from authenticated registries. It's built to be simple and straightforward, perfect for integrating into a `chart-proxy` Kubernetes deployment to point Kustomize towards.

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

### Docker Configuration Override

The Docker container includes a [placeholder config](examples/config/config.yaml). Replace it with your own:

```bash
# Mount your config file
docker run -v /path/to/your/config.yaml:/home/chart-fetcher/config.yaml michielvha/chart-fetcher:latest

# Or use command line flag
docker run -v /path/to/your/config.yaml:/app/config.yaml michielvha/chart-fetcher:latest --config /app/config.yaml

# Or use environment variable
docker run -v /path/to/your/config.yaml:/app/config.yaml -e CONFIG_PATH=/app/config.yaml michielvha/chart-fetcher:latest
```

For a kubernetes example, check [here](examples/manifests/chart-proxy/deployment.yaml)

# upcoming features

- **Application:**
  - proper output on run, add customization


[![Go Doc](https://pkg.go.dev/badge/github.com/michielvha/chart-fetcher.svg)](https://pkg.go.dev/github.com/michielvha/chart-fetcher)
[![license](https://img.shields.io/github/license/michielvha/chart-fetcher.svg?style=flat-square)](LICENSE)

[Go Report Card]: https://goreportcard.com/badge/github.com/michielvha/chart-fetcher
[go-report-card]: https://goreportcard.com/report/github.com/michielvha/chart-fetcher
[CodeQL]: https://github.com/michielvha/chart-fetcher/actions/workflows/github-code-scanning/codeql/badge.svg?branch=main
[code-ql]: https://github.com/michielvha/chart-fetcher/actions/workflows/github-code-scanning/codeql
