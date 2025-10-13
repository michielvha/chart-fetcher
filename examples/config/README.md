## 📝 Configuration Details

### chart-fetcher Configuration

The configuration supports:

**OCI Registries:**
```yaml
- url: "ghcr.io/myorg/charts"
  is_oci: true
  charts:
    - name: "mychart"
      version: "1.0.0"
```

**Legacy Helm Repositories:**
```yaml
- url: "https://charts.example.com"
  username_env: "REG1_USERNAME"
  password_env: "REG1_PASSWORD"
  charts:
    - name: "mychart"
      version: "1.0.0"
```
