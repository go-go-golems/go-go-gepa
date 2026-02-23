const plugins = require("./lib/gepa_plugin_contract");
const common = require("./lib/gepa_optimizer_common");

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

      const instruction = common.getCandidateText(
        candidate,
        "prompt",
        "Answer the question. Respond with only the final answer.",
      );

      const prompt = `${instruction}\n\nQuestion: ${String(example.question || "")}\nFinal answer:`;
      const got = common.runUserPrompt(ctx, options, prompt);
      const scored = common.exactMatchScore(example.answer, got);

      return {
        score: scored.score,
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

    function componentSideInfo(input) {
      const inObj = (input && typeof input === "object") ? input : {};
      const paramKey = common.toTrimmedString(inObj.paramKey) || "prompt";
      const fallback = (typeof inObj.default === "string") ? inObj.default : "";
      return `Component: ${paramKey}\n\n${fallback}`;
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
