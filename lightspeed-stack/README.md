# Lightspeed Stack

This directory contains a [Lightspeed Stack](https://github.com/lightspeed-core/lightspeed-stack) configuration
files to get it running for the purposes of this project.

## Quickstart

1. Run the [obs-mcp](../obs-mcp/README.md) MCP server
1. Clone https://github.com/lightspeed-core/lightspeed-stack
2. Copy the `lightspeed-stack.yaml` and `run.yaml` from this dir into the lightspeed-stack repository dir.
3. Follow the [Lightspeed Stack README](https://github.com/lightspeed-core/lightspeed-stack?tab=readme-ov-file#set-llm-provider-and-model) to configure the providers
5. `uv sync`
6. `make run`

### Model configuration

To change the model to be used by the chat service, configure `default_model/default_provider` inside `lightspeed-stack.yaml`
