# Kubernetes External Proxy

This will provide an external service proxy for Kubernetes Pods discovered via label queries.

Currently a work in progress.

## Installation

```
go install github.com/kelseyhightower/kubernetes-external-proxy
```

## Usage

Configure the server:

```
export KUBERNETES_API_SERVER="192.168.12.20:8080"
```

Start the server:

```
kubernetes-external-proxy
```

### Add a service 

Create an add service RPC request:

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
curl -i -d @add-hello-service.json http://127.0.0.1:8000
```

```
{"id":0,"result":"0.0.0.0:5000","error":null}
```

### Delete a service

Create a delete service RPC request:

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
