# AGENTS.md - Cluster Architecture & Operational Directives

## 1. System Overview
You are operating as the Central Hub in a distributed remote development cluster. Your primary objective is to maintain a lean context window and remain highly responsive. 

Heavy, long-running, or resource-intensive tasks (e.g., E2E testing, security scanning, database migrations, headless browser operations) are strictly offloaded to dedicated remote MCP (Model Context Protocol) Nodes.

**Architecture Model:**
* **Central Hub (You):** Orchestrates logic, writes code, and delegates tasks.
* **Sub-Agents:** Spawned by you to manage asynchronous waiting and log monitoring.
* **Remote MCP Nodes (WebDev, Android, Security, DB):** Isolated execution environments that handle physical/heavy execution.

## 2. The "Fire & Redirect" Execution Pattern
When a user requests a long-running test or task, **you must never execute it directly in the main thread.** You must use the following pattern:

1. **Delegate:** Spawn a sub-agent specifically tasked with overseeing the remote execution.
2. **Trigger:** The sub-agent invokes the required macro-tool on the relevant remote MCP Node (e.g., `webdev_run_e2e_tests`).
3. **Fire & Redirect (Node Side):** The remote MCP node initiates the task as a background process (`nohup`, detached container, etc.). It redirects all `stdout` and `stderr` to a specific log file located on a shared network mount accessible by the Hub. 
4. **Immediate Return:** The remote MCP tool immediately returns a JSON response containing the local path to the log file on the shared mount.
5. **Monitor:** The sub-agent passively tails/reads this local log file. It looks for predefined completion strings (e.g., `TEST SUITE COMPLETED`, `EXIT CODE: 0`) to determine the task's status.
6. **Report:** Once the sub-agent detects completion or failure, it parses the log, extracts the relevant stack trace or summary, and reports back to the main thread.

## 3. Remote MCP Node Blueprint & Tool Design
If you are tasked with creating, modifying, or interacting with an MCP Node within this cluster, adhere to these strict constraints:

* **No Micro-Tools:** MCP nodes must expose consolidated "macro-tools" (e.g., `trigger_playwright_suite`) rather than granular actions (e.g., `click_button`, `open_browser`).
* **Strict Namespacing:** All tool names must be prefixed with their node designation to prevent routing confusion (e.g., `android_run_emulator_tests`, `snyk_run_sast_scan`).
* **Verbose Background Logging:** The actual execution scripts (`Makefiles` or `npm scripts`) running on the nodes must output highly verbose logs. Because the MCP tool connection returns immediately, these logs written to the shared mount are the *only* way the Hub understands what is happening.
* **Statelessness:** MCP nodes do not retain state between tool calls. All artifacts (screenshots, coverage reports, test DB dumps) must be written directly to the shared network mounts.

## 4. Agent Rules of Engagement
* **DO NOT** attempt to wait synchronously for a remote test to finish.
* **DO NOT** pull entire raw log files into the main conversation context. Rely on the sub-agent to extract and summarize only the failing tests or critical errors.
* **DO** use the designated sub-agent to inform the user that a task has been offloaded (e.g., *"I've dispatched the Android tests to the remote node. I will notify you when the results are written to the shared mount."*)
* **DO** prioritize reading local artifact mounts over requesting data back over the MCP JSON-RPC channel.

## 5. Node Roster & Primary Functions
* **WebDev Node:** Headless Playwright/Storybook execution. Output: JSON summaries, PNG screenshots (via mounts).
* **Android Node:** Headless emulator testing. Output: Logcat dumps, APK build states.
* **Security Node:** Snyk/SAST codebase scanning. Output: Vulnerability reports.
* **Database Node:** Testcontainers, ephemeral Postgres/Redis state generation. Output: Connection strings, schema validation logs.
