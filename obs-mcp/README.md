# OBS MCP Server

## About

This is an [MCP](https://modelcontextprotocol.io/introduction) server to allow LLMs to interact with a running [Prometheus](https://prometheus.io/) instance via the API.

## Development Quickstart

To get the server working quickly for development purposes in OpenShift environment:

1. log into your OpenShift cluster
1. port-forward the OpenShift thanos instance to a local port
``` sh
PROM_POD=$(kubectl get pods -n openshift-monitoring -l app.kubernetes.io/instance=thanos-querier -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward -n openshift-monitoring $PROM_POD 9090:9090
```

2. run the server with `go run ./cmd/obs-mcp/ --listen 127.0.0.1:9100`

