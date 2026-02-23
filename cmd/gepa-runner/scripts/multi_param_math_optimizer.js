const plugins = require("geppetto/plugins");
const common = require("./lib/gepa_optimizer_common");

module.exports = plugins.defineOptimizerPlugin({
  apiVersion: plugins.OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "example.multi_param_math",
  name: "Example: Multi-param math optimizer",

  create(ctx) {
    const dataset = [
      { question: "14+5", answer: "19" },
      { question: "9*9", answer: "81" },
      { question: "56/8", answer: "7" },
      { question: "43-17", answer: "26" },
    ];

    let roundRobin = 0;

    function datasetFn() {
      return dataset;
    }

    function initialCandidate() {
      return {
        prompt: "Solve arithmetic carefully.",
        planner_prompt: "Plan the operation mentally before answering.",
        critic_prompt: "Double-check the arithmetic result before returning.",
      };
    }

    function composePrompt(candidate, question) {
      const planner = common.getCandidateText(candidate, "planner_prompt", "Plan the operation.");
      const base = common.getCandidateText(candidate, "prompt", "Solve arithmetic.");
      const critic = common.getCandidateText(candidate, "critic_prompt", "Verify the final value.");

      return [
        planner,
        base,
        critic,
        "",
        `Question: ${String(question || "")}`,
        "Return only the final numeric answer.",
      ].join("\n");
    }

    function evaluate(input, options) {
      const inObj = (input && typeof input === "object") ? input : {};
      const candidate = (inObj.candidate && typeof inObj.candidate === "object") ? inObj.candidate : {};
      const example = (inObj.example && typeof inObj.example === "object") ? inObj.example : {};

      const prompt = composePrompt(candidate, example.question);
      const got = common.runUserPrompt(ctx, options, prompt);
      const scored = common.exactMatchScore(example.answer, got);

      const feedbackByComponent = {
        prompt: scored.ok
          ? "Base task framing is working."
          : "Base prompt should stress strict arithmetic correctness and numeric-only output.",
        planner_prompt: scored.ok
          ? "Planning guidance seems adequate."
          : "Planning guidance should encourage identifying operation type first.",
        critic_prompt: scored.ok
          ? "Verification guidance appears effective."
          : "Critic guidance should require explicit self-check before returning.",
      };

      const traceByComponent = {
        prompt: {
          used: candidate.prompt,
          question: example.question,
          got: scored.got,
          expected: scored.expected,
        },
        planner_prompt: {
          used: candidate.planner_prompt,
          note: "Planner guidance precedes base prompt.",
        },
        critic_prompt: {
          used: candidate.critic_prompt,
          note: "Critic guidance follows base prompt.",
        },
      };

      return {
        score: scored.score,
        objectiveScores: {
          accuracy: scored.score,
          brevity: -common.toTrimmedString(candidate.prompt).length,
        },
        output: { text: scored.got },
        feedback: feedbackByComponent,
        trace: traceByComponent,
      };
    }

    function selectComponents(input) {
      const inObj = (input && typeof input === "object") ? input : {};
      const available = Array.isArray(inObj.availableKeys) ? inObj.availableKeys.slice() : [];
      if (available.length === 0) {
        return [];
      }

      if (inObj.operation === "merge") {
        const preferred = ["prompt", "critic_prompt", "planner_prompt"];
        return preferred.filter((k) => available.includes(k)).slice(0, 1);
      }

      const idx = roundRobin % available.length;
      roundRobin += 1;
      return [available[idx]];
    }

    function componentSideInfo(input) {
      const inObj = (input && typeof input === "object") ? input : {};
      const key = common.toTrimmedString(inObj.paramKey);
      const fallback = (typeof inObj.default === "string") ? inObj.default : "";

      const rubricByKey = {
        prompt: "Rubric: emphasize strict numeric output and arithmetic reliability.",
        planner_prompt: "Rubric: encourage concise plan before answering.",
        critic_prompt: "Rubric: enforce self-check for arithmetic slips.",
      };

      const rubric = rubricByKey[key] || "Rubric: improve the selected component.";
      return `${rubric}\n\n${fallback}`;
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
