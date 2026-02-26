const { defineOptimizerPlugin, OPTIMIZER_PLUGIN_API_VERSION } = require("gepa/plugins");

module.exports = defineOptimizerPlugin({
  apiVersion: OPTIMIZER_PLUGIN_API_VERSION,
  kind: "optimizer",
  id: "examples.gepa04.candidate-stream",
  name: "GEPA-04 Candidate Stream Demo",
  registryIdentifier: "local.examples",

  create() {
    return {
      run(input, options) {
        return Promise.resolve().then(() => {
          if (options && typeof options.emitEvent === "function") {
            options.emitEvent({
              type: "candidate-start",
              level: "info",
              data: {
                input_preview: String((input && input.text) || "").slice(0, 80),
              },
            });
          }
          if (options && options.events && typeof options.events.emit === "function") {
            options.events.emit({
              type: "candidate-progress",
              message: "running async candidate pipeline",
            });
          }

          return {
            output: {
              summary: "async candidate run completed",
              echoed_text: String((input && input.text) || ""),
            },
            metadata: {
              mode: "promise",
              plugin_id: "examples.gepa04.candidate-stream",
            },
          };
        });
      },
    };
  },
});
