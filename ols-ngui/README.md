# OLS & NGUI Integration

## Quickstart

### Prerequisities

1. Login to Openshift Cluster

    `oc login ...`

2. Run Openshift MCP Server with YAML output

    `npx kubernetes-mcp-server@latest --port 8081 --list-output yaml`

3. Run NGUI

   Use `quay.io/lkrzyzanek/next_gen_ui_mcp:dev` image

   ```sh
   export OPENAI_API_KEY="sk-..."
   podman run --rm -it -p 9200:9200 \
      --env MCP_PORT="9200" \
      --env NGUI_MODEL="gpt-4o" \
      --env NGUI_PROVIDER_API_KEY=$OPENAI_API_KEY \
      quay.io/lkrzyzanek/next_gen_ui_mcp:dev
   ``` 

    Or from git source: `python libs/next_gen_ui_mcp --provider langchain --model gpt-4o-mini  --port 9200 --transport streamable-http`

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