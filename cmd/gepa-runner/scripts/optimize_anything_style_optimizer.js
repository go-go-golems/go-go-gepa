const plugins = require("gepa/plugins");
const common = require("./lib/gepa_optimizer_common");

module.exports = plugins.defineOptimizerPlugin({
  apiVersion: plugins.OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "example.optimize_anything_style",
  name: "Example: Optimize-anything style adapter",

  create(ctx) {
    const COMPONENTS = [
      {
        key: "task_prompt",
        role: "core-task",
        defaultText: "Decide if the statement is TRUE or FALSE.",
        rubric: "Keep classification instructions precise.",
      },
      {
        key: "reasoning_prompt",
        role: "strategy",
        defaultText: "Reason briefly before deciding, then output only TRUE or FALSE.",
        rubric: "Reasoning guidance should reduce label errors.",
      },
      {
        key: "output_contract",
        role: "format",
        defaultText: "Output exactly one token: TRUE or FALSE.",
        rubric: "Output contract must be strict and minimal.",
      },
    ];

    const dataset = [
      { statement: "A triangle has three sides.", answer: "TRUE" },
      { statement: "2 is an odd number.", answer: "FALSE" },
      { statement: "Water freezes at 0C at sea level.", answer: "TRUE" },
      { statement: "The Earth has two moons.", answer: "FALSE" },
    ];

    let rr = 0;

    function datasetFn() {
      return dataset;
    }

    function initialCandidate() {
      const candidate = {};
      COMPONENTS.forEach((c) => {
        candidate[c.key] = c.defaultText;
      });
      return candidate;
    }

    function buildPrompt(candidate, statement) {
      return [
        common.getCandidateText(candidate, "task_prompt", "Classify statement truth."),
        common.getCandidateText(candidate, "reasoning_prompt", "Reason briefly."),
        common.getCandidateText(candidate, "output_contract", "Output TRUE or FALSE only."),
        "",
        `Statement: ${String(statement || "")}`,
      ].join("\n");
    }

    function evaluate(input, options) {
      const inObj = (input && typeof input === "object") ? input : {};
      const candidate = (inObj.candidate && typeof inObj.candidate === "object") ? inObj.candidate : {};
      const example = (inObj.example && typeof inObj.example === "object") ? inObj.example : {};

      const prompt = buildPrompt(candidate, example.statement);
      const got = common.runUserPrompt(ctx, options, prompt);
      const scored = common.exactMatchScore(example.answer, got.toUpperCase());

      const feedback = {};
      const trace = {};
      COMPONENTS.forEach((c) => {
        feedback[c.key] = scored.ok
          ? `${c.role} guidance is currently effective.`
          : `${c.role} guidance should reduce mismatch between expected and produced label.`;
        trace[c.key] = {
          role: c.role,
          used: candidate[c.key],
          expected: scored.expected,
          got: scored.got,
        };
      });

      return {
        score: scored.score,
        objectiveScores: {
          accuracy: scored.score,
          brevity: -prompt.length,
        },
        output: { label: scored.got },
        feedback,
        trace,
      };
    }

    function selectComponents(input) {
      const inObj = (input && typeof input === "object") ? input : {};
      const available = Array.isArray(inObj.availableKeys) ? inObj.availableKeys.slice() : [];
      if (available.length === 0) {
        return [];
      }

      // In merge mode, prefer output contract first to stabilize format consistency.
      if (inObj.operation === "merge" && available.includes("output_contract")) {
        return ["output_contract"];
      }

      const idx = rr % available.length;
      rr += 1;
      return [available[idx]];
    }

    function componentSideInfo(input) {
      const inObj = (input && typeof input === "object") ? input : {};
      const key = common.toTrimmedString(inObj.paramKey);
      const fallback = (typeof inObj.default === "string") ? inObj.default : "";

      const cfg = COMPONENTS.find((c) => c.key === key);
      if (!cfg) {
        return fallback;
      }

      return [
        `Component role: ${cfg.role}`,
        `Rubric: ${cfg.rubric}`,
        "",
        fallback,
      ].join("\n");
    }

    function merge(input, options) {
      return common.mergeWithLLM(ctx, options, input);
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
