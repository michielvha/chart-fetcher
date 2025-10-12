# Chart Proxy - chart-fetcher + Nginx Helm Repository

This deployment creates a local, unauthenticated Helm repository using chart-fetcher to pull charts from authenticated registries and nginx to serve them.

## 🎯 Purpose

This solves the problem where **ArgoCD + Kustomize** cannot pull Helm charts from authenticated registries. The chart-proxy:

1. **chart-fetcher** (init container) - Pulls charts from authenticated OCI/legacy registries
2. **Helm** (init container) - Generates the Helm repository index
3. **Nginx** (main container) - Serves charts as an unauthenticated HTTP Helm repository

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│           chart-proxy Pod                    │
├─────────────────────────────────────────────┤
│  Init Container 1: chart-fetcher             │
│  - Reads config from ConfigMap               │
│  - Authenticates to registries (via secrets)│
│  - Pulls .tgz files to /charts volume        │
├─────────────────────────────────────────────┤
│  Init Container 2: Helm                      │
│  - Generates index.yaml for charts          │
├─────────────────────────────────────────────┤
│  Container: Nginx                            │
│  - Serves /charts as HTTP repository         │
│  - No authentication required                │
└─────────────────────────────────────────────┘
         │
         ▼
   Kustomize can now fetch charts via HTTP
```

## 📦 Components

- `namespace.yaml` - Creates the `chart-proxy` namespace
- `configmap-chartfetch.yaml` - Configuration for which charts to pull
- `configmap-nginx.yaml` - Nginx server configuration
- `secret-registry-credentials.yaml` - Registry authentication credentials
- `deployment.yaml` - Main deployment with chart-fetcher + nginx
- `service.yaml` - ClusterIP service exposing nginx
- `kustomization.yaml` - Kustomize manifest

## 🚀 Quick Start

### 1. Update Configuration

Edit `configmap-chartfetch.yaml` to specify which charts to pull:

```yaml
data:
  config.yaml: |
    registries:
      - url: "quay.io/my-org/charts"
        is_oci: true
        charts:
          - name: "my-chart"
            version: "1.0.0"
```

### 2. Add Registry Credentials (if needed)

If your registries require authentication, update the secret:

```bash
kubectl create secret generic registry-credentials \
  --namespace=chart-proxy \
  --from-literal=username='myuser' \
  --from-literal=password='mypassword' \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 3. Deploy with kubectl

```bash
kubectl apply -k examples/manifests/chart-proxy/
```

### 4. Deploy with Kustomize

```bash
kustomize build examples/manifests/chart-proxy/ | kubectl apply -f -
```

## 🔧 Usage in ArgoCD/Kustomize

Once deployed, you can reference charts in your Kustomize manifests:

```yaml
# kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

helmCharts:
  - name: my-chart
    repo: http://nginx-service.chart-proxy.svc.cluster.local
    version: 1.0.0
    releaseName: my-release
    namespace: default
```

## 🔍 Verification

### Check if the deployment is running

```bash
kubectl get pods -n chart-proxy
```

Expected output:
```
NAME                          READY   STATUS    RESTARTS   AGE
chart-proxy-xxxxxxxxxx-xxxxx  1/1     Running   0          1m
```

### View pulled charts

```bash
kubectl exec -n chart-proxy deployment/chart-proxy -- ls -la /usr/share/nginx/html
```

### Check the Helm repository index

```bash
kubectl exec -n chart-proxy deployment/chart-proxy -- cat /usr/share/nginx/html/index.yaml
```

### Port-forward and test locally

```bash
kubectl port-forward -n chart-proxy svc/nginx-service 8080:80
```

Then browse to http://localhost:8080 or:

```bash
curl http://localhost:8080/index.yaml
helm repo add local-proxy http://localhost:8080
helm search repo local-proxy
```

## 📝 Configuration Details

### chart-fetcher Configuration

The `configmap-chartfetch.yaml` supports:

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

### Environment Variables

Set in `secret-registry-credentials`:
- `REG1_USERNAME` - Registry username
- `REG1_PASSWORD` - Registry password

You can add more environment variables in the deployment for multiple registries (REG2_USERNAME, REG2_PASSWORD, etc.)

## 🔄 Updating Charts

To pull new charts or versions:

1. Update the ConfigMap with new chart definitions
2. Delete the pod to trigger a restart:
   ```bash
   kubectl delete pod -n chart-proxy -l app=chart-proxy
   ```

The init containers will run again and pull the updated chart list.

## 🛠️ Troubleshooting

### chart-fetcher logs

```bash
kubectl logs -n chart-proxy deployment/chart-proxy -c chartfetch
```

### Helm repo index logs

```bash
kubectl logs -n chart-proxy deployment/chart-proxy -c helm-repo-index
```

### Nginx logs

```bash
kubectl logs -n chart-proxy deployment/chart-proxy -c nginx
```

### Common Issues

1. **Charts not appearing**: Check chart-fetcher logs for authentication errors
2. **index.yaml missing**: Check helm-repo-index container logs
3. **Cannot connect**: Ensure service is created and pod is running

## 🎨 Customization

### Change the chart-fetcher image

Update the image in `deployment.yaml`:
```yaml
- name: chartfetch
  image: your-registry/chartfetch:tag
```

### Add resource limits

Edit `deployment.yaml` to add resources:
```yaml
resources:
  limits:
    memory: "128Mi"
    cpu: "100m"
  requests:
    memory: "64Mi"
    cpu: "50m"
```

### Expose externally

Change service type to `LoadBalancer` or add an `Ingress`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
  namespace: chart-proxy
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
  selector:
    app: chart-proxy
```

## 📚 Related Documentation

- [chart-fetcher README](../../../readme.md)
- [Kustomize Helm Charts](https://kubectl.docs.kubernetes.io/references/kustomize/builtins/#_helmchartinflationgenerator_)
- [Helm Repository Structure](https://helm.sh/docs/topics/chart_repository/)

## 🎯 Use Cases

1. **ArgoCD + Kustomize Integration** - Pull charts from authenticated registries
2. **Internal Chart Mirror** - Cache external charts for air-gapped environments
3. **Chart Aggregation** - Combine charts from multiple registries into one
4. **Development** - Test chart deployments without pushing to a registry

## 🔐 Security Notes

- The nginx service serves charts **without authentication** by design
- Use Kubernetes NetworkPolicies to restrict access if needed
- Keep registry credentials in secrets, never in ConfigMaps
- Consider using External Secrets Operator for credential management
