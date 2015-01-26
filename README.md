# Kubernetes External Proxy

This will provide an external service proxy for Kubernetes Pods discovered via label queries.

Currently a work in progress.

## Usage

```
export KUBERNETES_API_SERVER="192.168.12.20:8080"
```

```
kubernetes-external-proxy
```

### Add a service 

Create a service request

```
{
    "method": "ServiceManager.Add",
    "params":[{
        "id": "hello",
        "selector": {
            "environment": "production"
        },
        "containerPort": "80",
        "protocol": "tcp",
        "port": "5000"
    }],
    "id": 0
}
```

```
curl -i -d @hello-service.json http://127.0.0.1:8000
```

```
{"id":0,"result":"0.0.0.0:5000","error":null}
```

### Delete a service

```
{
    "method": "ServiceManager.Del",
    "params":["hello"],
    "id": 0
}
```

```
curl -i -d @delete-hello-service.json http://127.0.0.1:8000
```

```
{"id":0,"result":true,"error":null}
```
