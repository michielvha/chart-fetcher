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
      "url": "oci://registry1.example.com",
      "username_env": "REGISTRY1_USERNAME",
      "password_env": "REGISTRY1_PASSWORD",
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
