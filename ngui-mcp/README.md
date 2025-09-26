# Next Gen UI MCP Server

## About

This is an [MCP](https://modelcontextprotocol.io/introduction) server to allow LLMs to interact with a running [Next Gen UI](https://github.com/RedHat-UX/next-gen-ui-agent) instance via the API.

## Development Quickstart

### Installation

```sh
cd ngui-mcp
python3 -m venv .venv && source .venv/bin/activate
pip install -U next_gen_ui_mcp langchain_openai
```

### Run server

See examples how to run MCP server: https://github.com/RedHat-UX/next-gen-ui-agent/tree/main/libs/next_gen_ui_mcp

#### Open AI

Set API Key `OPENAI_API_KEY` env variable

```sh
export OPENAI_API_KEY="sk...."
source .venv/bin/activate
python -m next_gen_ui_mcp --port 9200 --transport streamable-http --provider langchain --model gpt-4o-mini
```

Expected output:
```
2025-09-25 13:57:13,582 - __main__ - INFO - Starting Next Gen UI MCP Server with streamable-http transport
2025-09-25 13:57:13,583 - __main__ - INFO - Using component system: json
2025-09-25 13:57:13,583 - __main__ - INFO - Using LangChain inference with model gpt-4o-mini
2025-09-25 13:57:13,949 - __main__ - INFO - Server running on http://127.0.0.1:9200/mcp
INFO:     Started server process [22570]
INFO:     Waiting for application startup.
2025-09-25 13:57:13,953 - mcp.server.streamable_http_manager - INFO - StreamableHTTP session manager started
INFO:     Application startup complete.
INFO:     Uvicorn running on http://127.0.0.1:9200 (Press CTRL+C to quit)
```

#### Ollama


```sh
source .venv/bin/activate
python -m next_gen_ui_mcp --port 9200 --transport streamable-http --provider langchain --model llama3.2 --base-url http://localhost:11434/v1 --api-key ollama
```

## Test MCP Server

Run inspector e.g. `npx @modelcontextprotocol/inspector@latest` and configure:

* Transport Type: `Streamable HTTP`
* URL: `http://127.0.0.1:9200/mcp`

and Connect.
Go to Tools and perform `List Tools`

## TS Test Client

Go to [test-ts-client](./test-ts-client/) directory.
