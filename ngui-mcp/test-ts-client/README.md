# Next Gen UI MCP Server - Test TS Client

## Install dependencies

```sh
npm install
```

## Run Example

```sh
npm run start
```

Expected output:


```sh
============= CLIENT INIT (http://127.0.0.1:9200/mcp)=================
============= CLIENT SUCCESS =================
============= TOOLS =================
{
  tools: [
    {
      name: 'generate_ui',
      description: 'Generate UI components from user prompt and input data.\n' +
        '\n' +
        'This tool can use either external inference providers or MCP sampling capabilities.\n' +
        'When external inference is provided, it uses that directly. Otherwise, it creates\n' +
        "an InferenceBase using MCP sampling to leverage the client's LLM.\n" +
        '\n' +
        'Args:\n' +
        "    user_prompt: User's request or prompt describing what UI to generate\n" +
        "    input_data: List of input data items with 'id' and 'data' keys\n" +
        '    ctx: MCP context providing access to sampling capabilities\n' +
        '\n' +
        'Returns:\n' +
        '    List of rendered UI components ready for display\n',
      inputSchema: [Object],
      outputSchema: [Object]
    }
  ]
}
============= TOOL_CALL generate_ui(user_prompt, input_data) =================
============= TOOL_RESULT generate_ui =================
{
  "component": "table",
  "id": "some_id",
  "title": "Namespaces",
  "fields": [
    {
      "name": "Name",
      "data_path": "$..namespaces[*].metadata.name",
      "data": [
        "default",
        "kube-node-lease",
        "kube-public",
        "kube-system",
        "openshift",
        "openshift-apiserver",
....
```