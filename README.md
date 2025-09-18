## Prerequisites

Before starting, ensure you have the following:

- A working Lightspeed Core-based server with the capability to integrate the MCP server located in the `obs-mcp` directory of this project.
- Access to a model capable of tool calling. This project was tested with `gpt-4o-mini`.
- An environment where both Node.js (version 20 or higher) and Golang are available. Using `nvm` (Node Version Manager) and `gvm` (Go Version Manager) is recommended for managing multiple versions.
- Access to an OpenShift Container Platform (OCP) cluster with the monitoring plugin installed.

## Getting Started

Follow these steps to get up and running:

1. Set up the obs-mcp server. For details, see the [obs-mcp README](./obs-mcp/README.md).
2. Once the server is running, connect it to your Lightspeed Core (LSC) instance.
3. Start the console UI: in the `dynamic-plugin` package, run `yarn start-console`.
4. Start the UI plugin by running `yarn start` in the `dynamic-plugin` directory.
5. Open your browser and navigate to `http://localhost:9000/genie`.
