import { z } from 'zod';
import { spawnDetached } from '../lib/spawn.js';
import { DATA_ROOT, projectName } from '../lib/paths.js';
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';
import os from 'node:os';
export const playwrightSchema = z.object({
    project_path: z.string().min(1),
    spec: z.string().optional(),
    test_name: z.string().optional(),
    tag: z.string().optional(),
});
export async function runPlaywright(params) {
    const { project_path, spec, test_name, tag } = params;
    if (test_name && tag) {
        throw new Error('test_name and tag are mutually exclusive');
    }
    const proj = projectName(project_path);
    const grepValue = tag ? `@${tag}` : test_name;
    const pwArgs = ['playwright', 'test', '--reporter=line', '--screenshot=only-on-failure'];
    if (spec)
        pwArgs.push(spec);
    if (grepValue)
        pwArgs.push('--grep', grepValue);
    const label = spec ?? tag ?? test_name ?? 'all';
    const wrapper = join(os.tmpdir(), `couchdev-playwright-${Date.now()}.sh`);
    writeFileSync(wrapper, `#!/bin/sh
PID=$$
LOG_DIR="${DATA_ROOT}/logs/${proj}/$PID"
SS_DIR="${DATA_ROOT}/screenshots/${proj}/$PID"
mkdir -p "$LOG_DIR" "$SS_DIR"
LOG="$LOG_DIR/playwright.log"
exec >> "$LOG" 2>&1
cd ${JSON.stringify(project_path)}
PLAYWRIGHT_HTML_OUTPUT_DIR="$SS_DIR" npx ${pwArgs.map(a => JSON.stringify(a)).join(' ')}
CODE=$?
if [ $CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $CODE
`, { mode: 0o755 });
    const pid = await spawnDetached('/bin/sh', [wrapper], project_path, '/dev/null');
    return { pid };
}
