# initialize Go project using Azure-Devops

```golang
go mod init dev.azure.com/bnl-ms/AzureFoundation/charthost
```


### **Example Run Scenarios**

| **Scenario**                               | **Result**                              |
|--------------------------------------------|-----------------------------------------|
| No input                                   | Uses `./config.json` as default.        |
| `--config /path/to/config.json`            | Uses `/path/to/config.json`.            |
| `CHARTHOST_CONFIG_PATH=/path/from/env.json`| Uses `/path/from/env.json`.             |
| Both `--config` and `CHARTHOST_CONFIG_PATH`| Environment variable takes precedence.  |

### **Example Configuration File**


````json
{
  "registries": [
    {
      "url": "engiebnlms.jfrog.io/artifactory/prd-helm-virtual/",
      "username_env": "REG1_USERNAME",
      "password_env": "REG1_PASSWORD",
      "charts": [
        { "name": "chart1", "version": "latest" },
        { "name": "chart2", "version": "2.1.0" }
      ]
    },
    {
      "url": "oci://registry2.example.com",
      "charts": [
        { "name": "chart3", "version": "latest" }
      ]
    }
  ]
}
````
## Troubleshooting

### login with helm to ACR

````shell
helm registry login mslocalfoundationacr.azurecr.io --username $REG1_USERNAME --password-stdin $REG1_PASSWORD
````

### Login to Artifactory

You seem to need to add repo to be able to pull from Artifactory

````shell 

````

## Changelog

**1.0.0 - Initial release**
- Added support for json & yaml configuration files
- Added support for environment variables
- Added support for command line arguments
  - `--config`: Specify a custom configuration file path. Defaults to `config.json`
  - `--outputPath`: Specify a custom output path for the charts. Default to `charts/`
- Added support for OCI registries through 
  - `Login`: To conform to the new OCI standard we must use something similar to helm login.
  - `PullOCIChart`: Pulling charts via the SDK using the `h.RegistryClient` method.
- Added support for Legacy repositories through: 
  - `AddRepo`: Legacy repositories need you to add the repo to the helm client.
  - `FetchRepoIndex`: Repo index file needs to be fetched by the binary to be able to search the repo.
  - `EnsureRepoFileExists`: check if the initial repostories.yaml file exists, if not create it.
  - `PullLegacyChart`: use the old method via http to pull the chart by reading the index file.
