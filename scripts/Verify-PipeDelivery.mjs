import assert from "node:assert/strict";
import {
    execFileSync,
    spawn,
} from "node:child_process";
import { once } from "node:events";
import fs from "node:fs";
import path from "node:path";
import {
    fileURLToPath,
    pathToFileURL,
} from "node:url";

const root = fileURLToPath(new URL("../", import.meta.url));
const compiler = path.resolve(process.argv[2] ?? path.join(root, "built/local", process.platform === "win32" ? "tsgo.exe" : "tsgo"));
const expectedVersion = "7.0.2-pipe.0";
const run = (exe, args) => execFileSync(exe, args, { cwd: root, encoding: "utf8", windowsHide: true, timeout: 30_000 }).trim();
assert.equal(run(compiler, ["--version"]), `Version ${expectedVersion}`);
assert.ok(fs.existsSync(path.join(path.dirname(compiler), "lib.es5.d.ts")), "library declarations must accompany the compiler");

const example = path.join(root, "examples/pipe.ts");
for (const target of ["es2015", "es2022", "esnext"]) {
    const out = path.join(root, "built/verification/pipe-delivery", target);
    run(compiler, ["--ignoreConfig", "--strict", "--declaration", "--sourceMap", "--target", target, "--module", "esnext", "--outDir", out, example]);
    fs.writeFileSync(path.join(out, "package.json"), JSON.stringify({ type: "module" }));
    assert.equal(run(process.execPath, [path.join(out, "pipe.js")]), "pipe examples ok");
    const declarations = fs.readFileSync(path.join(out, "pipe.d.ts"), "utf8");
    assert.match(declarations, /export declare const length: number/);
    assert.match(declarations, /sum\(a\?: number, b\?: number\): number/);
    console.log(`${target}: strict compile, Node execution and declarations passed`);
}

const server = spawn(compiler, ["--lsp", "--stdio"], { cwd: root, windowsHide: true });
const exited = once(server, "exit");
// Attach rejection handling immediately, including the spawn-failure case.
exited.catch(() => {});
let stderr = "";
server.stderr.setEncoding("utf8").on("data", chunk => {
    stderr += chunk;
});
const pending = new Map();
let nextId = 0;
let buffer = Buffer.alloc(0);

function send(message) {
    const body = Buffer.from(JSON.stringify({ jsonrpc: "2.0", ...message }));
    server.stdin.write(Buffer.concat([Buffer.from(`Content-Length: ${body.length}\r\n\r\n`), body]));
}

function fail(error) {
    for (const entry of pending.values()) {
        clearTimeout(entry.timer);
        entry.reject(error);
    }
    pending.clear();
}

server.on("error", fail);
server.stdin.on("error", fail);
server.on("exit", code => fail(new Error(`LSP exited (${code}): ${stderr}`)));
server.stdout.on("data", chunk => {
    buffer = Buffer.concat([buffer, chunk]);
    try {
        while (true) {
            const headerEnd = buffer.indexOf("\r\n\r\n");
            if (headerEnd === -1) return;
            const length = /^Content-Length:\s*(\d+)\s*$/im.exec(buffer.subarray(0, headerEnd).toString());
            assert.ok(length, "LSP response must have a Content-Length header");
            const end = headerEnd + 4 + Number(length[1]);
            if (buffer.length < end) return;
            const message = JSON.parse(buffer.subarray(headerEnd + 4, end).toString());
            buffer = buffer.subarray(end);
            if (message.method && message.id !== undefined) {
                send({ id: message.id, result: null });
            }
            else if (pending.has(message.id)) {
                const entry = pending.get(message.id);
                pending.delete(message.id);
                clearTimeout(entry.timer);
                if (message.error) entry.reject(new Error(JSON.stringify(message.error)));
                else entry.resolve(message.result);
            }
        }
    }
    catch (error) {
        fail(error);
        server.kill();
    }
});

function request(method, params) {
    const id = ++nextId;
    return new Promise((resolve, reject) => {
        const timer = setTimeout(() => {
            pending.delete(id);
            reject(new Error(`LSP request timed out: ${method}\n${stderr}`));
        }, 30_000);
        pending.set(id, { resolve, reject, timer });
        send({ id, method, params });
    });
}

try {
    const initialized = await request("initialize", {
        processId: null,
        rootUri: pathToFileURL(root).href,
        capabilities: { textDocument: { hover: { contentFormat: ["plaintext"] } } },
    });
    assert.equal(initialized.serverInfo.version, expectedVersion);
    assert.ok(initialized.capabilities.hoverProvider);
    send({ method: "initialized", params: {} });
    const source = fs.readFileSync(example, "utf8");
    const uri = pathToFileURL(example).href;
    send({ method: "textDocument/didOpen", params: { textDocument: { uri, languageId: "typescript", version: 1, text: source } } });
    const prefix = source.slice(0, source.indexOf("const length") + "const ".length);
    const lines = prefix.split("\n");
    const hover = await request("textDocument/hover", {
        textDocument: { uri },
        position: { line: lines.length - 1, character: lines.at(-1).length + 1 },
    });
    assert.match(JSON.stringify(hover?.contents), /length: number/);
    await request("shutdown");
    send({ method: "exit" });
    const timer = setTimeout(() => server.kill(), 5_000);
    try {
        const [code] = await exited;
        assert.equal(code, 0, stderr);
    }
    finally {
        clearTimeout(timer);
    }
    console.log(`LSP: ${expectedVersion}, pipe hover and clean shutdown passed`);
}
finally {
    if (server.exitCode === null) server.kill();
    fail(new Error("LSP verification finished"));
}
