import { readFile } from 'fs/promises';
import { Client } from "@modelcontextprotocol/sdk/client/index.js";
import { StreamableHTTPClientTransport } from "@modelcontextprotocol/sdk/client/streamableHttp.js";
import { NotificationSchema, RequestSchema, ResultSchema } from '@modelcontextprotocol/sdk/types.js';
import * as z from "zod";

const ngui_mcp_url = "http://127.0.0.1:9200/mcp"
const transport = new StreamableHTTPClientTransport(new URL(ngui_mcp_url));


// Custom schemas
// Does not work :-(
const CustomRequestSchema = RequestSchema.extend({})
const CustomNotificationSchema = NotificationSchema.extend({})
const CustomResultSchema = ResultSchema.extend({
    id: z.string(),
    content: z.string(),
    name: z.string(),
})
// Type aliases
type CustomRequest = z.infer<typeof CustomRequestSchema>
type CustomNotification = z.infer<typeof CustomNotificationSchema>
type CustomResult = z.infer<typeof CustomResultSchema>

// Create typed client
const client = new Client<CustomRequest, CustomNotification, CustomResult>({
    name: "genie-client",
    version: "1.0.0"
})

type InputDataType = {
    id: string,
    data: string,
}

type NguiOutputType = {
    id: string,
    content: string,
    name: string,
}



async function generate_ui(user_prompt: string, input_data: Array<InputDataType>) {

    console.log("============= TOOL_CALL generate_ui(user_prompt, input_data) =================")

    // Call a tool
    const result = await client.callTool({
        name: "generate_ui",
        arguments: {
            user_prompt,
            input_data
        }
    });
    console.log("============= TOOL_RESULT generate_ui =================")
    // console.log(result);
    if (result.isError) throw Error("Error during calling tool" + result)

    const ngui_ui_response: NguiOutputType = result.structuredContent.result[0]
    const ui_block = JSON.parse(ngui_ui_response.content)
    console.log(JSON.stringify(ui_block, null, 2));
}

async function run() {
    console.log("============= CLIENT INIT (%s)=================", ngui_mcp_url)
    await client.connect(transport);
    console.log("============= CLIENT SUCCESS =================")

    // List tools
    const tools = await client.listTools()
    console.log("============= TOOLS =================")
    console.log(tools)

    const namespaces_all = await readFile("kube_namespaces_all_mock.json", "utf8");
    const user_prompt = "What are my all namespaces?"
    await generate_ui(user_prompt, [{ id: "some_id", data: namespaces_all }])
}

run()
    .catch(err => { console.error(err) })
    .then(client.close).catch(e => process.exit(1));
