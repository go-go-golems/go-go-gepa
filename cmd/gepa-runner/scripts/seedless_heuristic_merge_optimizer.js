const plugins = require("geppetto/plugins");
const common = require("./lib/gepa_optimizer_common");

module.exports = plugins.defineOptimizerPlugin({
  apiVersion: plugins.OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "example.seedless_heuristic_merge",
  name: "Example: Seedless with heuristic merge",

  create(ctx) {
    const dataset = [
      { question: "3+4", answer: "7" },
      { question: "18-6", answer: "12" },
      { question: "11*2", answer: "22" },
      { question: "24/6", answer: "4" },
    ];

    function datasetFn() {
      return dataset;
    }

    // Designed for --seedless mode.
    function initialCandidate() {
      return {
        prompt: "Solve the arithmetic question and return only the final answer.",
      };
    }

    function evaluate(input, options) {
      const inObj = (input && typeof input === "object") ? input : {};
      const candidate = (inObj.candidate && typeof inObj.candidate === "object") ? inObj.candidate : {};
      const example = (inObj.example && typeof inObj.example === "object") ? inObj.example : {};

      const instruction = common.getCandidateText(candidate, "prompt", "Return only the final answer.");
      const prompt = `${instruction}\n\nQuestion: ${String(example.question || "")}\nAnswer:`;

      const got = common.runUserPrompt(ctx, options, prompt);
      const scored = common.exactMatchScore(example.answer, got);

      return {
        score: scored.score,
        objectiveScores: {
          accuracy: scored.score,
          prompt_cost: -instruction.length,
        },
        output: { text: scored.got },
        feedback: scored.feedback,
      };
    }

    function selectComponents(input) {
      const inObj = (input && typeof input === "object") ? input : {};
      const available = Array.isArray(inObj.availableKeys) ? inObj.availableKeys : [];
      if (available.includes("prompt")) {
        return ["prompt"];
      }
      return available.slice(0, 1);
    }

    // Merge strategy without additional model calls.
    function merge(input) {
      const inObj = (input && typeof input === "object") ? input : {};
      const a = common.toTrimmedString(inObj.paramA);
      const b = common.toTrimmedString(inObj.paramB);

      let merged = a.length >= b.length ? a : b;
      const mustHave = [
        "Return only the final answer.",
        "Do not include explanations.",
      ];
      mustHave.forEach((rule) => {
        if (!merged.toLowerCase().includes(rule.toLowerCase())) {
          merged = `${merged} ${rule}`.trim();
        }
      });
      return merged;
    }

    return {
      dataset: datasetFn,
      initialCandidate,
      evaluate,
      selectComponents,
      merge,
    };
  },
});
