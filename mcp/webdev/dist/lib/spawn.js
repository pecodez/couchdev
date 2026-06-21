import { spawn } from 'node:child_process';
import fs from 'node:fs';
export async function spawnDetached(cmd, args, cwd, logPath) {
    const out = logPath === '/dev/null' ? 'ignore' : fs.openSync(logPath, 'w');
    const child = spawn(cmd, args, {
        cwd,
        stdio: ['ignore', out === 'ignore' ? 'ignore' : out, out === 'ignore' ? 'ignore' : out],
        detached: true,
    });
    child.unref();
    if (typeof out === 'number')
        fs.closeSync(out);
    if (child.pid === undefined) {
        throw new Error(`Failed to spawn ${cmd}`);
    }
    return child.pid;
}
