import { execFileSync } from "node:child_process";
import { lstatSync, readFileSync, realpathSync } from "node:fs";
import { isAbsolute, join, relative, sep } from "node:path";
import { parseArgs } from "node:util";

const SOURCE_PATTERNS = ["*.js", "*.jsx", "*.mjs", "*.cjs", "*.ts", "*.tsx", "*.mts", "*.cts"];

const EXCLUDED = ["node_modules/", "dist/", "build/", "coverage/"];

const ALLOWED = [
  /^\/\/\/\s*<reference\b/,
  /^\/[/*]\s*@ts-(check|nocheck|expect-error|ignore)\b/,
  /^\/[/*]\s*(eslint|oxlint)-(disable|enable)/,
  /^\/[/*]\s*(biome-ignore|prettier-ignore)\b/,
  /^\/[/*]\s*(istanbul|c8|v8) ignore\b/,
  /^\/[/*]\*?\s*@(vitest|jest)-environment\b/,
  /^\/[/*]\*?\s*@jsx(ImportSource|Runtime|Frag)?\b/,
  /^\/\*\s*[#@]__(PURE|NO_SIDE_EFFECTS)__\s*\*\/$/,
  /^\/\*\s*(webpack[A-Z]\w*|@vite-ignore)\b/,
  /^\/\/ workaround: \S+$/,
];

const JSX_WORKAROUND = /^\/\* workaround: \S+ \*\/$/;

const LICENSE_START = /^\/(\/|\*[*!]?)[\s*]*(Copyright\b|\([cC]\)\s|©|SPDX-License-Identifier:|@license\b|@preserve\b)/;
const LICENSE_BANNER = /^\/\*![\s\S]*(\blicen[cs]e[ds]?\b|\bcopyright\b|SPDX-License-Identifier|\([cC]\)|©)/i;
const LICENSE_TAG = /^\/\*[*!]?[^\n]*?(\n[ \t*]*)?@(license|preserve)\b|\n[ \t*]*@(license|preserve)\b/;
const TASK_NOTE = /^\/[/*]\**\s*(TODO|FIXME|HACK|XXX|BUG)\b/i;
const BLANK_LINE = /(\r\n|\n|\r)[ \t]*(\r\n|\n|\r)/;

const JSDOC_TYPES = /^\/\*\*[\s\S]*@(type|typedef|param\s*\{|returns?\s*\{|template|satisfies|import)\b/;
const JS_FILES = /\.(js|jsx|mjs|cjs)$/;

const PUBLIC_API_SYMBOLS = new Map();

const DECISIONS_PATH = "_support/docs/decisions/";

class Failure extends Error {}

function git(root, args) {
  try {
    return execFileSync("git", ["-C", root, ...args], {
      encoding: "utf8",
      maxBuffer: 256 * 1024 * 1024,
      stdio: ["ignore", "pipe", "pipe"],
    });
  } catch (error) {
    const detail = error.stderr ? String(error.stderr).trim() : error.message;
    throw new Failure(`git ${args[0]} failed: ${detail}`);
  }
}

function nulSeparated(output) {
  return output.split("\0").filter(Boolean);
}

function excluded(file) {
  return EXCLUDED.some((entry) => (entry.endsWith("/") ? file.startsWith(entry) : file === entry));
}

function pluginsFor(file) {
  if (/\.(ts|mts|cts)$/.test(file)) return ["typescript", "decorators-legacy"];
  if (/\.tsx$/.test(file)) return ["typescript", "jsx", "decorators-legacy"];
  return ["jsx", "decorators-legacy"];
}

function parseSource(parser, file, text) {
  try {
    return parser.parse(text, {
      sourceType: "unambiguous",
      plugins: pluginsFor(file),
      tokens: true,
      errorRecovery: false,
    });
  } catch (error) {
    throw new Failure(`${file}: could not parse: ${error.message}`);
  }
}

function isDirective(file, text) {
  return ALLOWED.some((pattern) => pattern.test(text)) || (JS_FILES.test(file) && JSDOC_TYPES.test(text));
}

function publicApiDocStarts(program, file, text) {
  const starts = new Set();
  const symbols = PUBLIC_API_SYMBOLS.get(file);
  if (!symbols) return starts;
  for (const statement of program.body) {
    const declaration = statement.type === "ExportNamedDeclaration" || statement.type === "ExportDefaultDeclaration"
      ? statement.declaration
      : statement;
    if (!declaration) continue;
    const binding = declaration.type === "VariableDeclaration" && declaration.declarations.length === 1
      ? declaration.declarations[0].id
      : declaration.id;
    if (!binding || binding.type !== "Identifier" || !symbols.has(`${declaration.type}:${binding.name}`)) continue;
    const doc = statement.leadingComments?.at(-1);
    if (doc?.type === "CommentBlock" && text.startsWith("/**", doc.start) &&
        !text.slice(doc.end, statement.start).trim()) starts.add(doc.start);
  }
  return starts;
}

function comments(parser, file, text, publicApiPath = file) {
  const ast = parseSource(parser, file, text);
  const publicDocs = publicApiDocStarts(ast.program, publicApiPath, text);
  const containers = [];
  walk(ast.program, (node) => {
    if (node.type === "JSXExpressionContainer") containers.push([node.start, node.end]);
  });
  const codeStart = ast.program.body.length ? ast.program.body[0].start : text.length;
  const header = new Set();
  let previous = null;
  for (const comment of ast.comments) {
    if (comment.end > codeStart) break;
    if (previous && BLANK_LINE.test(text.slice(previous.end, comment.start))) break;
    const raw = text.slice(comment.start, comment.end);
    if (!previous && !LICENSE_START.test(raw) && !LICENSE_BANNER.test(raw)) break;
    if (!TASK_NOTE.test(raw)) header.add(comment);
    previous = comment;
  }
  return ast.comments.map((comment) => {
    const raw = text.slice(comment.start, comment.end);
    const lineStart = text.lastIndexOf("\n", comment.start - 1) + 1;
    const lineEnd = text.indexOf("\n", comment.end);
    const before = text.slice(lineStart, comment.start);
    const after = text.slice(comment.end, lineEnd === -1 ? text.length : lineEnd);
    return {
      line: comment.loc.start.line,
      column: comment.loc.start.column + 1,
      start: comment.start,
      end: comment.end,
      fullLine: before.trim() === "" && after.trim() === "",
      publicApiDoc: publicDocs.has(comment.start),
      directive:
        isDirective(file, raw) ||
        header.has(comment) ||
        (comment.type === "CommentBlock" && LICENSE_TAG.test(raw)) ||
        (JSX_WORKAROUND.test(raw) &&
          containers.some(([start, end]) => start <= comment.start && comment.end <= end)),
      text: raw,
    };
  });
}

function normalized(text) {
  return text
    .split(/\r\n|\n|\r/)
    .map((line) => line.trim())
    .join("\n");
}

function readSource(root, file) {
  const path = join(root, file);
  let stats;
  try {
    stats = lstatSync(path);
  } catch {
    throw new Failure(`${file}: cannot read`);
  }
  if (!stats.isFile()) return null;
  const real = realpathSync(path);
  const inside = relative(realpathSync(root), real);
  if (inside === ".." || inside.startsWith(`..${sep}`) || isAbsolute(inside)) return null;
  return readFileSync(real, "utf8");
}

function shown(text) {
  return [...text]
    .map((character) => {
      const code = character.charCodeAt(0);
      return code < 32 || code === 127 ? `\\x${code.toString(16).padStart(2, "0")}` : character;
    })
    .join("");
}

function resolveBase(root, base) {
  if (base === undefined) {
    try {
      return git(root, ["merge-base", "HEAD", "refs/remotes/origin/HEAD"]).trim();
    } catch {
      throw new Failure("no --base given and origin/HEAD is unknown; pass --base <rev>");
    }
  }
  try {
    return git(root, ["rev-parse", "--verify", "--end-of-options", `${base}^{commit}`]).trim();
  } catch {
    throw new Failure(`base ${base} does not resolve to a commit`);
  }
}

function changedFiles(root, base) {
  const fields = nulSeparated(
    git(root, [
      "diff",
      "--no-ext-diff",
      "--no-textconv",
      "--find-renames",
      "--name-status",
      "-z",
      base,
      "--",
      ...SOURCE_PATTERNS,
    ]),
  );
  const changed = new Map();
  for (let index = 0; index < fields.length; ) {
    const status = fields[index];
    if (status.startsWith("R") || status.startsWith("C")) {
      changed.set(fields[index + 2], { before: fields[index + 1], status: status[0] });
      index += 3;
    } else {
      const file = fields[index + 1];
      changed.set(file, { before: status === "A" ? null : file, status });
      index += 2;
    }
  }
  const untracked = nulSeparated(
    git(root, ["ls-files", "-z", "--others", "--exclude-standard", "--", ...SOURCE_PATTERNS]),
  );
  for (const file of untracked) changed.set(file, { before: null, status: "A" });
  return [...changed]
    .filter(([file, { before }]) => !excluded(file) || (before !== null && before !== file && !excluded(before)))
    .map(([file, { before, status }]) => ({ file, before, status }));
}

function excludedChanges(root, base) {
  if (!EXCLUDED.length) return [];
  const listed = nulSeparated(
    git(root, [
      "diff",
      "--no-ext-diff",
      "--no-textconv",
      "--no-renames",
      "--ignore-submodules=dirty",
      "--name-only",
      "-z",
      base,
      "--",
      ...EXCLUDED.map((entry) => `:(top,literal)${entry}`),
    ]),
  );
  return listed.filter(excluded);
}

function baseText(root, base, file) {
  return git(root, ["cat-file", "blob", `${base}:${file}`]);
}

function lint(parser, root, base) {
  let findings = 0;
  for (const { file, before, status } of changedFiles(root, base)) {
    if (status === "D" || excluded(file)) continue;
    const text = readSource(root, file);
    if (text === null) continue;
    const remaining = new Map();
    if (before !== null) {
      for (const comment of comments(parser, before, baseText(root, base, before), file)) {
        if (comment.publicApiDoc) continue;
        const key = normalized(comment.text);
        remaining.set(key, (remaining.get(key) ?? 0) + 1);
      }
    }
    for (const comment of comments(parser, file, text)) {
      if (comment.directive || comment.publicApiDoc) continue;
      const key = normalized(comment.text);
      const left = remaining.get(key) ?? 0;
      if (left > 0) {
        remaining.set(key, left - 1);
        continue;
      }
      console.error(shown(`${file}:${comment.line}:${comment.column}: no-new-comments`));
      findings++;
    }
  }
  if (findings) {
    console.error(
      `${findings} new comment(s). Record a lasting decision in ${DECISIONS_PATH} instead.`,
    );
  }
  return findings ? 1 : 0;
}

function inventory(parser, root) {
  const files = new Set(
    nulSeparated(
      git(root, [
        "ls-files",
        "-z",
        "--cached",
        "--others",
        "--exclude-standard",
        "--",
        ...SOURCE_PATTERNS,
      ]),
    ),
  );
  const deleted = new Set(
    nulSeparated(git(root, ["ls-files", "-z", "--deleted", "--", ...SOURCE_PATTERNS])),
  );
  for (const file of files) {
    if (excluded(file) || deleted.has(file)) continue;
    const text = readSource(root, file);
    if (text === null) continue;
    for (const comment of comments(parser, file, text)) {
      console.log(JSON.stringify({ file, ...comment }));
    }
  }
  return 0;
}

function walk(node, visit) {
  if (!node || typeof node !== "object") return;
  if (Array.isArray(node)) {
    for (const child of node) walk(child, visit);
    return;
  }
  if (typeof node.type === "string") visit(node);
  for (const [key, value] of Object.entries(node)) {
    if (key !== "loc" && key !== "comments" && key !== "tokens" && key !== "extra") walk(value, visit);
  }
}

function codeTokens(parser, file, text) {
  const ast = parseSource(parser, file, text);
  const emptyContainers = [];
  walk(ast.program, (node) => {
    if (node.type === "JSXExpressionContainer" && node.expression?.type === "JSXEmptyExpression") {
      emptyContainers.push([node.start, node.end]);
    }
  });
  const tokens = [];
  let previousEnd = null;
  for (const token of ast.tokens) {
    const label = typeof token.type === "string" ? token.type : token.type.label;
    if (label === "CommentLine" || label === "CommentBlock") continue;
    const container = emptyContainers.find(([start, end]) => token.start >= start && token.end <= end);
    if (container) {
      previousEnd = container[1];
      continue;
    }
    const value = text.slice(token.start, token.end);
    const breakBefore = previousEnd !== null && /[\n\r]/.test(text.slice(previousEnd, token.start));
    previousEnd = token.end;
    if (label === "jsxText") {
      const previous = tokens.at(-1);
      if (previous?.label === "jsxText") {
        previous.value += value;
      } else {
        tokens.push({ label, value, breakBefore: false });
      }
      continue;
    }
    tokens.push({ label, value, breakBefore: tokens.at(-1)?.label === "jsxText" ? false : breakBefore });
  }
  return tokens
    .map((token) =>
      token.label === "jsxText" ? { ...token, value: token.value.replace(/\s+/g, " ").trim() } : token,
    )
    .filter((token) => !(token.label === "jsxText" && token.value === ""))
    .map((token) => `${token.breakBefore ? "\\n " : ""}${token.label} ${token.value}`);
}

function sameCode(parser, root, base) {
  let findings = 0;
  const reported = new Set();
  for (const file of excludedChanges(root, base)) {
    console.error(shown(`${file}: changed under an excluded path`));
    reported.add(file);
    findings++;
  }
  for (const { file, before, status } of changedFiles(root, base)) {
    if (reported.has(file) || reported.has(before)) {
      if (before !== null && before !== file && !excluded(before)) {
        console.error(shown(`${file}: moved from ${before}`));
        findings++;
      }
      continue;
    }
    if (status === "D") {
      console.error(shown(`${file}: deleted`));
      findings++;
      continue;
    }
    if (before !== null && before !== file) {
      console.error(shown(`${file}: moved from ${before}`));
      findings++;
      continue;
    }
    const text = readSource(root, file);
    if (text === null) {
      console.error(shown(`${file}: not a regular file`));
      findings++;
      continue;
    }
    const now = codeTokens(parser, file, text);
    if (before === null) {
      if (now.length) {
        console.error(shown(`${file}: added with code`));
        findings++;
      }
      continue;
    }
    const old = codeTokens(parser, before, baseText(root, base, before));
    const index = old.findIndex((token, position) => token !== now[position]);
    if (index !== -1 || old.length !== now.length) {
      const at = index === -1 ? Math.min(old.length, now.length) : index;
      console.error(shown(`${file}: code changed at token ${at + 1}: ${old[at] ?? "(end)"} -> ${now[at] ?? "(end)"}`));
      findings++;
    }
  }
  return findings ? 1 : 0;
}

async function main() {
  const { values } = parseArgs({
    options: {
      base: { type: "string" },
      inventory: { type: "boolean" },
      "same-code": { type: "boolean" },
    },
  });
  let parser;
  try {
    parser = await import("@babel/parser");
  } catch {
    throw new Failure("@babel/parser is not installed; add it as a devDependency");
  }
  const root = git(process.cwd(), ["rev-parse", "--show-toplevel"]).trimEnd();
  if (values.inventory) {
    if (values.base !== undefined || values["same-code"]) {
      throw new Failure("--inventory takes no other option");
    }
    return inventory(parser, root);
  }
  const base = resolveBase(root, values.base);
  return values["same-code"] ? sameCode(parser, root, base) : lint(parser, root, base);
}

try {
  process.exitCode = await main();
} catch (error) {
  console.error(shown(error instanceof Failure ? error.message : String(error?.stack ?? error)));
  process.exitCode = 2;
}
