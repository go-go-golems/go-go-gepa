const { defineOptimizerPlugin, OPTIMIZER_PLUGIN_API_VERSION } = require("gepa/plugins");

module.exports = defineOptimizerPlugin({
  apiVersion: OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "examples.candidate-run.smoke",
  name: "Candidate Run Smoke Plugin",
  registryIdentifier: "local.examples",
  create() {
    return {
      evaluate(input) {
        const candidate = (input && typeof input === "object" && input.candidate && typeof input.candidate === "object")
          ? input.candidate
          : {};
        const example = (input && typeof input === "object" && input.example && typeof input.example === "object")
          ? input.example
          : {};
        const prompt = String(candidate.prompt || "");
        const question = String(example.question || "");
        return {
          score: 1,
          objectives: { score: 1 },
          output: { prompt, question },
          feedback: "ok"
        };
      },
      run(input, options) {
        const inObj = (input && typeof input === "object") ? input : {};
        const candidate = (options && typeof options === "object" && options.candidate && typeof options.candidate === "object")
          ? options.candidate
          : {};
        const prompt = String(candidate.prompt || "");
        const planner = String(candidate.planner_prompt || "");
        const question = String(inObj.question || "");
        const expected = String(inObj.answer || "");
        return {
          output: {
            prompt,
            planner_prompt: planner,
            question,
            expected_answer: expected,
            composed_instruction: prompt + "\n\n" + planner
          },
          metadata: {
            mode: "candidate-run",
            plugin: "examples.candidate-run.smoke"
          }
        };
      }
    };
  }
});
