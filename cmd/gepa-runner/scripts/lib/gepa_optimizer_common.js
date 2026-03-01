const gp = require("geppetto");

function toTrimmedString(v) {
  return String(v == null ? "" : v).trim();
}

function resolveAssistantText(out) {
  const blocks = (out && Array.isArray(out.blocks)) ? out.blocks : [];
  return blocks
    .filter((b) => b && b.payload && typeof b.payload.text === "string")
    .filter((b) => b.kind === gp.consts.BlockKind.LLM_TEXT || b.kind === "assistant")
    .map((b) => toTrimmedString(b.payload.text))
    .join("\n")
    .trim();
}

function extractTripleBacktickBlock(s) {
  const text = toTrimmedString(s);
  const m = text.match(/```(?:[a-zA-Z0-9_-]+)?\s*([\s\S]*?)\s*```/);
  if (m && m[1]) {
    return toTrimmedString(m[1]);
  }
  return text;
}

function normalizeEngineOptions(ctx, options) {
  const merged = (options && typeof options === "object") ? options : {};
  const ctxOpts = (ctx && typeof ctx === "object") ? ctx : {};

  const engineOptions = (merged.engineOptions && typeof merged.engineOptions === "object")
    ? merged.engineOptions
    : ((ctxOpts.engineOptions && typeof ctxOpts.engineOptions === "object") ? ctxOpts.engineOptions : null);

  const profile = toTrimmedString(merged.profile)
    || toTrimmedString(ctxOpts.profile)
    || "";

  return { profile, engineOptions };
}

function createEngine(ctx, options) {
  const resolved = normalizeEngineOptions(ctx, options);
  if (resolved.engineOptions && Object.keys(resolved.engineOptions).length > 0) {
    return gp.engines.fromConfig(resolved.engineOptions);
  }
  return gp.engines.fromProfile(resolved.profile || "", {});
}

function runUserPrompt(ctx, options, prompt) {
  const engine = createEngine(ctx, options);
  const builder = gp.createBuilder().withEngine(engine);
  const session = builder.buildSession();

  const seed = gp.turns.newTurn({
    blocks: [gp.turns.newUserBlock(String(prompt || ""))],
  });

  const out = session.run(seed, {});
  return resolveAssistantText(out);
}

function getCandidateText(candidate, key, fallback) {
  const c = (candidate && typeof candidate === "object") ? candidate : {};
  if (typeof c[key] === "string" && c[key].trim()) {
    return c[key].trim();
  }
  if (key !== "prompt" && typeof c.prompt === "string" && c.prompt.trim()) {
    return c.prompt.trim();
  }
  return toTrimmedString(fallback || "");
}

function exactMatchScore(expected, got) {
  const e = toTrimmedString(expected);
  const g = toTrimmedString(got);
  const ok = e !== "" && g === e;
  return {
    score: ok ? 1.0 : 0.0,
    ok,
    expected: e,
    got: g,
    feedback: ok ? "Correct." : `Expected "${e}" but got "${g}".`,
  };
}

function defaultMergePrompt(input) {
  const inObj = (input && typeof input === "object") ? input : {};
  const paramA = toTrimmedString(inObj.paramA);
  const paramB = toTrimmedString(inObj.paramB);
  const sideA = String(inObj.sideInfoA || "");
  const sideB = String(inObj.sideInfoB || "");

  return [
    "You are optimizing an instruction prompt.",
    "Merge the two candidates below into one improved instruction.",
    "Keep it concise and unambiguous.",
    "",
    "Instruction A:",
    "```",
    paramA,
    "```",
    "",
    "Instruction B:",
    "```",
    paramB,
    "```",
    "",
    "Examples and feedback for A:",
    "```",
    sideA,
    "```",
    "",
    "Examples and feedback for B:",
    "```",
    sideB,
    "```",
    "",
    "Return only the merged instruction in triple backticks.",
  ].join("\n");
}

function mergeWithLLM(ctx, options, input) {
  const prompt = defaultMergePrompt(input);
  const text = runUserPrompt(ctx, options, prompt);
  return extractTripleBacktickBlock(text);
}

module.exports = {
  gp,
  toTrimmedString,
  resolveAssistantText,
  extractTripleBacktickBlock,
  normalizeEngineOptions,
  createEngine,
  runUserPrompt,
  getCandidateText,
  exactMatchScore,
  defaultMergePrompt,
  mergeWithLLM,
};
