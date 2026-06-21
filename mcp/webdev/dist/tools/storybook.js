import { z } from 'zod';
import { spawnDetached } from '../lib/spawn.js';
import { DATA_ROOT, projectName } from '../lib/paths.js';
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';
import os from 'node:os';
export const storybookSchema = z.object({
    project_path: z.string().min(1),
});
export async function runStorybook(params) {
    const { project_path } = params;
    const proj = projectName(project_path);
    const wrapper = join(os.tmpdir(), `couchdev-storybook-${Date.now()}.sh`);
    writeFileSync(wrapper, `#!/bin/sh
PID=$$
LOG_DIR="${DATA_ROOT}/logs/${proj}/$PID"
mkdir -p "$LOG_DIR"
LOG="$LOG_DIR/storybook.log"
exec >> "$LOG" 2>&1
cd ${JSON.stringify(project_path)}
npx storybook test --headless
CODE=$?
if [ $CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $CODE
`, { mode: 0o755 });
    const pid = await spawnDetached('/bin/sh', [wrapper], project_path, '/dev/null');
    return { pid };
}
