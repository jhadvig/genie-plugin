# OBS MCP Server

An [MCP](https://modelcontextprotocol.io/introduction) server that enables LLMs to interact with Prometheus/Thanos instances via a dynamic, configurable API.

## Key Features

- **Dynamic Tool**: Single `prometheus_query` tool for all configured Prometheus API endpoints
- **Configurable Access**: Limit which endpoints are accessible (default or full mode)
- **PromQL Guardrails**: Query validation, length limits, and custom endpoint handlers
- **Documentation Resources**: Automatic access to official Prometheus docs from GitHub
- **Authentication**: Supports kubeconfig, service account, and header-based auth
- **Dev Mode**: Easy testing with kind clusters via `--dev` flag

## Quick Start

### OpenShift / Thanos Querier

```sh
# Login to cluster
oc login ...

# Run server (auto-discovers Thanos Querier)
go run ./cmd/obs-mcp/ \
  --listen 127.0.0.1:9100 \
  --auth-mode kubeconfig \
  --insecure
```

### Kind Cluster / Prometheus (Dev Mode)

```sh
# Setup port-forward
kubectl port-forward -n monitoring svc/prometheus-k8s 9090:9090

# Run server with dev mode (no authentication needed for localhost)
go run ./cmd/obs-mcp/ \
  --listen 127.0.0.1:9100 \
  --auth-mode none \
  --dev
```

### Port-Forward Alternative

```sh
# Port-forward Thanos Querier
PROM_POD=$(kubectl get pods -n openshift-monitoring \
  -l app.kubernetes.io/instance=thanos-querier \
  -o jsonpath="{.items[0].metadata.name}")
kubectl port-forward -n openshift-monitoring $PROM_POD 9090:9090

# Run server (no authentication needed for localhost)
PROMETHEUS_URL=http://localhost:9090 go run ./cmd/obs-mcp/ \
  --listen 127.0.0.1:9100 \
  --auth-mode none
```

## Configuration

### CLI Flags

```sh
--listen <addr>              # HTTP server address (required for HTTP mode)
--auth-mode <mode>          # Authentication: kubeconfig, serviceaccount, or header
--insecure                  # Skip TLS verification (default: true)
--enable-all-endpoints      # Enable all Prometheus endpoints (default: false)
--dev                       # Dev mode: target Prometheus in kind cluster instead of Thanos
```

### Endpoint Modes

**Default**
- `/api/v1/label/__name__/values` - List all metrics
- `/api/v1/query_range` - Execute range queries

**Full** (enable with `--enable-all-endpoints`, 9+ endpoints):
- Query: `/api/v1/query`, `/api/v1/query_range`
- Series: `/api/v1/series`, `/api/v1/labels`, `/api/v1/label/{name}/values`
- Status: `/api/v1/targets`, `/api/v1/alertmanagers`, `/api/v1/status/config`, `/api/v1/status/flags`

### Authentication Modes

**kubeconfig**: Uses bearer token from kubeconfig (requires `oc login` or `kubectl`)
**serviceaccount**: Reads token from `/var/run/secrets/kubernetes.io/serviceaccount/token`
**header**: Extracts bearer token from `Authorization` header in requests
**none**: No authentication (suitable for localhost port-forward scenarios)

### Environment Variables

```sh
PROMETHEUS_URL=<url>        # Override Prometheus URL (bypasses auto-discovery)
```

## Usage Example

Query Prometheus using the dynamic tool:

```json
{
  "tool": "prometheus_query",
  "arguments": {
    "endpoint": "/api/v1/query_range",
    "parameters": {
      "query": "rate(container_cpu_usage_seconds_total[5m])",
      "start": "2024-01-01T00:00:00Z",
      "end": "2024-01-01T01:00:00Z",
      "step": "15s"
    }
  }
}
```

List available metrics:

```json
{
  "tool": "prometheus_query",
  "arguments": {
    "endpoint": "/api/v1/label/__name__/values"
  }
}
```

## Architecture

### Dynamic Tool System

Single `prometheus_query` tool that:
- Accepts any configured Prometheus API endpoint path
- Validates endpoints against an allowlist (EndpointConfig)
- Applies PromQL guardrails for query endpoints
- Supports custom validation handlers per endpoint
- Supports authentication and TLS handling

### Guardrails

Built-in safety system:
- **PromQL Validation**: Query length limits (10KB max), syntax checks
- **Endpoint Validation**: Ensures only allowed endpoints are accessed
- **Custom Handlers**: Per-endpoint validation logic (e.g., time range limits)

### Resources

Prometheus documentation automatically fetched from GitHub:
- Querying basics and PromQL syntax
- Query operators and functions
- HTTP API reference
- Query examples

## Development

**Use `--dev` flag** to target Prometheus in a local kind cluster for easier testing.

### Project Structure

```
pkg/mcp/
├── endpoints.go          # Endpoint paths (constants) and configurations
├── dynamic_handler.go    # Single handler for all Prometheus queries
├── guardrails.go         # PromQL validation and safety checks
├── resources.go          # Prometheus documentation resources
├── server.go             # MCP server setup and tool registration
└── auth.go               # Authentication (kubeconfig/serviceaccount/header)

pkg/k8s/
└── client.go             # Kubernetes client and service discovery

pkg/prometheus/
└── client.go             # Prometheus API client wrapper
```

### Adding Custom Guardrails

Register custom validation for specific endpoints:

```go
// In server.go SetupUniversalTool()
guardrails.RegisterHandler(PathQueryRange,
  func(ctx context.Context, endpoint PrometheusEndpoint, params map[string]interface{}) error {
    // Custom validation (e.g., enforce max time range)
    return nil
  })
```

### Adding New Endpoints

Edit `pkg/mcp/endpoints.go`:

1. Add path constant:
```go
const PathYourEndpoint = "/api/v1/your_endpoint"
```

2. Add to `FullEndpointConfig()`:
```go
config.AllowedEndpoints[PathYourEndpoint] = PrometheusEndpoint{
  Path:               PathYourEndpoint,
  Description:        "Your endpoint description",
  Method:             "GET",
  RequiresGuardrails: false,
  Parameters: []EndpointParameter{
    {Name: "param1", Type: "string", Description: "...", Required: true},
  },
}
```

## Future Enhancements

- [ ] Rate limiting per endpoint
- [ ] Audit logging for all queries
- [ ] [Request hooks](https://github.com/mark3labs/mcp-go#request-hooks) for observability into request lifecycles
- [ ] Environmental awareness:
  - Detect cluster topology and adjust behavior (e.g., hypershift's METRICS_SET)
  - Detect collection profiles (or cluster capabilites) and avoid suggesting unavailable metrics
- [ ] E2E test using the `lightspeed-evaluation` framework
