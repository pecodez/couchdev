import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js'
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js'
import { storybookSchema, runStorybook } from './tools/storybook.js'
import { playwrightSchema, runPlaywright } from './tools/playwright.js'
import { buildSchema, runBuild } from './tools/build.js'
import { previewSchema, runPreview } from './tools/preview.js'

const server = new McpServer({
  name: 'webdev',
  version: '0.1.0',
})

server.tool(
  'webdev_run_storybook',
  'Run Storybook test runner headlessly against the project. Returns pid immediately; monitor DATA_ROOT/logs/<project>/<pid>/storybook.log for COUCHDEV:EXIT:0/1.',
  storybookSchema.shape,
  async (params) => {
    const result = await runStorybook(params)
    return { content: [{ type: 'text' as const, text: JSON.stringify(result) }] }
  },
)

server.tool(
  'webdev_run_playwright',
  'Run Playwright tests with optional spec file, test_name, or tag filter (tag and test_name are mutually exclusive). Returns pid immediately.',
  playwrightSchema.shape,
  async (params) => {
    const result = await runPlaywright(params)
    return { content: [{ type: 'text' as const, text: JSON.stringify(result) }] }
  },
)

server.tool(
  'webdev_build',
  'Run vite build --mode <mode> in the project. Returns pid immediately; monitor DATA_ROOT/logs/<project>/<pid>/build-<mode>.log.',
  buildSchema.shape,
  async (params) => {
    const result = await runBuild(params)
    return { content: [{ type: 'text' as const, text: JSON.stringify(result) }] }
  },
)

server.tool(
  'webdev_preview',
  'Build with vite, serve with vite preview, take a headless Chromium screenshot. Returns pid immediately; screenshot at DATA_ROOT/screenshots/<project>/<pid>/preview-<mode>.png.',
  previewSchema.shape,
  async (params) => {
    const result = await runPreview(params)
    return { content: [{ type: 'text' as const, text: JSON.stringify(result) }] }
  },
)

const transport = new StdioServerTransport()
await server.connect(transport)
