const gp = require("geppetto");
const plugins = require("geppetto/plugins");

function resolveAssistantText(out) {
  const blocks = (out && Array.isArray(out.blocks)) ? out.blocks : [];
  return blocks
    .filter((b) => b && b.payload && typeof b.payload.text === "string")
    .filter((b) => b.kind === gp.consts.BlockKind.LLM_TEXT || b.kind === "assistant")
    .map((b) => (b.payload.text || "").trim())
    .join("\n")
    .trim();
}

function extractTripleBacktickBlock(s) {
  const text = String(s || "");
  const m = text.match(/```(?:[a-zA-Z0-9_-]+)?\s*([\s\S]*?)\s*```/);
  if (m && m[1]) {
    return String(m[1]).trim();
  }
  return text.trim();
}

function normalizeEngineOptions(ctx, options) {
  const merged = (options && typeof options === "object") ? options : {};
  const ctxOpts = (ctx && typeof ctx === "object") ? ctx : {};
  const engineOptions = (merged.engineOptions && typeof merged.engineOptions === "object")
    ? merged.engineOptions
    : ((ctxOpts.engineOptions && typeof ctxOpts.engineOptions === "object") ? ctxOpts.engineOptions : null);

  // If no explicit config, fall back to profile.
  return {
    profile: (typeof merged.profile === "string" && merged.profile.trim())
      ? merged.profile.trim()
      : ((typeof ctxOpts.profile === "string" && ctxOpts.profile.trim()) ? ctxOpts.profile.trim() : ""),
    engineOptions,
  };
}

module.exports = plugins.defineOptimizerPlugin({
  apiVersion: plugins.OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "example.toy_math",
  name: "Example: Toy math accuracy",

  create(ctx) {
    const dataset = [
      { question: "2+2", answer: "4" },
      { question: "10-3", answer: "7" },
      { question: "6*7", answer: "42" },
      { question: "12/4", answer: "3" },
      { question: "9+8", answer: "17" },
      { question: "100-25", answer: "75" },
    ];

    function datasetFn() {
      return dataset;
    }

    function initialCandidate() {
      return {
        prompt: "Answer the question. Respond with only the final answer.",
      };
    }

    function evaluate(input, options) {
      const inObj = (input && typeof input === "object") ? input : {};
      const candidate = (inObj.candidate && typeof inObj.candidate === "object") ? inObj.candidate : {};
      const example = (inObj.example && typeof inObj.example === "object") ? inObj.example : {};

      const instruction = (typeof candidate.prompt === "string" && candidate.prompt.trim())
        ? candidate.prompt.trim()
        : "Answer the question. Respond with only the final answer.";

      const prompt = `${instruction}\n\nQuestion: ${String(example.question || "")}\nFinal answer:`;

      const resolved = normalizeEngineOptions(ctx, options);
      const engine = (resolved.engineOptions && Object.keys(resolved.engineOptions).length > 0)
        ? gp.engines.fromConfig(resolved.engineOptions)
        : gp.engines.fromProfile(resolved.profile || "", {});

      const builder = gp.createBuilder().withEngine(engine);
      const session = builder.buildSession();

      const seed = gp.turns.newTurn({
        blocks: [
          gp.turns.newUserBlock(prompt),
        ],
      });

      const out = session.run(seed, {});
      const text = resolveAssistantText(out);

      const expected = String(example.answer || "").trim();
      const got = String(text || "").trim();

      const ok = expected !== "" && got === expected;
      return {
        score: ok ? 1.0 : 0.0,
        output: { text: got },
        feedback: ok ? "Correct." : `Expected "${expected}" but got "${got}".`,
      };
    }

    // Optional: plugin-side component selection hook.
    // For this toy single-param plugin, always optimize "prompt" when available.
    function selectComponents(input, options) {
      const inObj = (input && typeof input === "object") ? input : {};
      const available = Array.isArray(inObj.availableKeys) ? inObj.availableKeys : [];
      if (available.includes("prompt")) {
        return ["prompt"];
      }
      return available.slice(0, 1);
    }

    // Optional: plugin-side side-info shaping hook.
    function componentSideInfo(input, options) {
      const inObj = (input && typeof input === "object") ? input : {};
      const fallback = (typeof inObj.default === "string") ? inObj.default : "";
      const paramKey = (typeof inObj.paramKey === "string" && inObj.paramKey.trim())
        ? inObj.paramKey.trim()
        : "prompt";
      return `Component: ${paramKey}\n\n${fallback}`;
    }

    // Optional: custom prompt merge (crossover).
    // If omitted, gepa-runner will use its built-in LLM merge template.
    function merge(input, options) {
      const inObj = (input && typeof input === "object") ? input : {};
      const candidateA = (inObj.candidateA && typeof inObj.candidateA === "object") ? inObj.candidateA : {};
      const candidateB = (inObj.candidateB && typeof inObj.candidateB === "object") ? inObj.candidateB : {};
      const paramKey = (typeof inObj.paramKey === "string" && inObj.paramKey.trim()) ? inObj.paramKey.trim() : "prompt";

      const a = (typeof inObj.paramA === "string" && inObj.paramA.trim())
        ? inObj.paramA.trim()
        : String(candidateA[paramKey] || candidateA.prompt || "").trim();
      const b = (typeof inObj.paramB === "string" && inObj.paramB.trim())
        ? inObj.paramB.trim()
        : String(candidateB[paramKey] || candidateB.prompt || "").trim();

      const sideA = (typeof inObj.sideInfoA === "string") ? inObj.sideInfoA : "";
      const sideB = (typeof inObj.sideInfoB === "string") ? inObj.sideInfoB : "";

      const prompt = [
        "You are optimizing an instruction prompt for a math QA assistant.",
        "Merge the two candidate instructions below into a single improved instruction.",
        "Keep it concise and specific.",
        "",
        "Instruction A:",
        "```",
        a,
        "```",
        "",
        "Instruction B:",
        "```",
        b,
        "```",
        "",
        "Examples & feedback for A:",
        "```",
        sideA,
        "```",
        "",
        "Examples & feedback for B:",
        "```",
        sideB,
        "```",
        "",
        "Return the merged instruction within ``` blocks.",
      ].join("\n");

      const resolved = normalizeEngineOptions(ctx, options);
      const engine = (resolved.engineOptions && Object.keys(resolved.engineOptions).length > 0)
        ? gp.engines.fromConfig(resolved.engineOptions)
        : gp.engines.fromProfile(resolved.profile || "", {});

      const builder = gp.createBuilder().withEngine(engine);
      const session = builder.buildSession();

      const seed = gp.turns.newTurn({
        blocks: [
          gp.turns.newUserBlock(prompt),
        ],
      });

      const out = session.run(seed, {});
      const text = resolveAssistantText(out);
      return extractTripleBacktickBlock(text);
    }

    return {
      dataset: datasetFn,
      initialCandidate,
      evaluate,
      selectComponents,
      componentSideInfo,
      merge,
    };
  },
});
