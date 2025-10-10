# OLS & NGUI Integration

## Quickstart

### Prerequisities

1. Login to Openshift Cluster

    `oc login ...`

2. Run Openshift MCP Server with YAML output

    `npx kubernetes-mcp-server@latest --port 8081 --list-output yaml --read-only --toolsets core`

3. Run NGUI

   ```sh
   export OPENAI_API_KEY="sk-..."
   podman run --rm -it -p 9200:9200 \
      -v $PWD/ols-ngui:/opt/app-root/config:z \
      --env MCP_PORT="9200" \
      --env NGUI_MODEL="gpt-4o-mini" \
      --env NGUI_PROVIDER_API_KEY=$OPENAI_API_KEY \
      --env NGUI_CONFIG_PATH="/opt/app-root/config/ngui_openshift_mcp_config.yaml" \
      quay.io/next-gen-ui/mcp
   ``` 

Or from git source:
    
```sh
PYTHONPATH=./libs python libs/next_gen_ui_mcp --provider langchain --model gpt-4o-mini  --port 9200 --transport streamable-http --config-path /Users/lkrzyzan/git/genie/genie-plugin/ols-ngui/ngui_openshift_mcp_config.yaml
```

### Start OLS

1. Clone repo https://github.com/lkrzyzanek/lightspeed-service, branch “ngui-mcp” and install deps.

    ```sh
    cd git/genie
    git clone https://github.com/lkrzyzanek/lightspeed-service.git
    git checkout ngui-mcp
    cd lightspeed-service
    make install-deps 
    ```

2. Copy `olsconfig.yaml`

    ```sh
    cp ../genie-plugin/olsconfig.yaml .
    ```

3. Run OLS

    ```sh
    export OPENAI_API_KEY="sk-..."
    pdm run python runner.py
    ```

## Test

```sh
curl --request POST \
  --url http://localhost:8080/v1/streaming_query \
  --header 'Content-Type: application/json' \
  --header 'User-Agent: insomnium/0.2.3-a' \
  --data '{
  "media_type": "application/json",
  "model": "gpt-4o-mini",
  "provider": "openai",
  "query": "what are my namespaces (and generate ui)?"
}'
```

You can change `"media_type": "application/json",` to `"media_type": "text/plain",`


## Conversation Examples

### Openshift Namespace to Pod

```
what are my namespaces, generate ui

what pods are running in openshift-lightspeed namespace, generate ui
    what pods are running in openshift-lightspeed namespace

tell me all details about pod lightspeed-app-server-8d87bd889-rhxm4, generate ui
	generate again the component about that pod

what is restart policy for that pod?
what is restart policy for that pod, generate ui
    ^ This is fully generated one card component
```

### Unknown data - Dashboards

```
what pods are running in openshift-lightspeed namespace, generate ui

what are my dashboards? generate ui
	what are my dashboards? Include all possible information, generate ui
```

### Perses

```
create a dashboard called Libor POC and add a widget showing me the CPU usage for the pods in my openshift-monitoring namespace over the last hour
```