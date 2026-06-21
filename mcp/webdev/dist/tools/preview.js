import { z } from 'zod';
import { spawnDetached } from '../lib/spawn.js';
import { DATA_ROOT, projectName } from '../lib/paths.js';
import { writeFileSync } from 'node:fs';
import { join } from 'node:path';
import os from 'node:os';
export const previewSchema = z.object({
    project_path: z.string().min(1),
    mode: z.string().min(1),
});
export async function runPreview(params) {
    const { project_path, mode } = params;
    const proj = projectName(project_path);
    const wrapper = join(os.tmpdir(), `couchdev-preview-${Date.now()}.sh`);
    writeFileSync(wrapper, `#!/bin/sh
PID=$$
LOG_DIR="${DATA_ROOT}/logs/${proj}/$PID"
SS_DIR="${DATA_ROOT}/screenshots/${proj}/$PID"
mkdir -p "$LOG_DIR" "$SS_DIR"
LOG="$LOG_DIR/preview-${mode}.log"
SS="$SS_DIR/preview-${mode}.png"
exec >> "$LOG" 2>&1

cd ${JSON.stringify(project_path)}

npx vite build --mode ${JSON.stringify(mode)}
if [ $? -ne 0 ]; then echo "COUCHDEV:EXIT:1"; exit 1; fi

# Start preview on a random free port
npx vite preview --port 0 --strictPort false &
PREVIEW_PID=$!
sleep 3

# Detect bound port
PORT=$(ss -tlnp 2>/dev/null | awk '/vite|node/{match($4, /:([0-9]+)$/, a); if(a[1]) print a[1]}' | head -1)
if [ -z "$PORT" ]; then PORT=4173; fi

node --input-type=module <<JS
import { chromium } from 'playwright';
const browser = await chromium.launch();
const page = await browser.newPage();
await page.goto('http://localhost:$PORT');
await page.waitForLoadState('networkidle');
await page.screenshot({ path: '$SS', fullPage: true });
await browser.close();
JS
SS_CODE=$?

kill $PREVIEW_PID 2>/dev/null
wait $PREVIEW_PID 2>/dev/null

if [ $SS_CODE -eq 0 ]; then echo "COUCHDEV:EXIT:0"; else echo "COUCHDEV:EXIT:1"; fi
exit $SS_CODE
`, { mode: 0o755 });
    const pid = await spawnDetached('/bin/sh', [wrapper], project_path, '/dev/null');
    return { pid };
}
