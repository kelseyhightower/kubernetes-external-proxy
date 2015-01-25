# Kubernetes External Proxy

This will provide an external service proxy for Kubernetes Pods discovered via label queries.

Currently a work in progress.

## Usage

```
export KUBERNETES_API_SERVER="http://192.168.12.20:8080"
```

```
kubernetes-external-proxy
```

```
curl -i -d @hello-service.json http://127.0.0.1:8000
```
