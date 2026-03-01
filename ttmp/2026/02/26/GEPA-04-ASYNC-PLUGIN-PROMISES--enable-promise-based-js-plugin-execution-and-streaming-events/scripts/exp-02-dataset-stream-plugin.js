const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "examples.gepa04.dataset-stream",
  name: "GEPA-04 Dataset Stream Demo",
  registryIdentifier: "local.examples",

  create() {
    return {
      generateOne(input, options) {
        const idx = Number((input && input.index) || 0);
        return Promise.resolve().then(() => {
          if (options && options.events && typeof options.events.emit === "function") {
            options.events.emit({
              type: "row-start",
              data: { index: idx },
            });
          }
          if (options && typeof options.emitEvent === "function") {
            options.emitEvent({
              type: "row-finish",
              message: `row-${idx} complete`,
            });
          }

          return {
            row: {
              id: `row-${idx}`,
              text: `Synthetic coaching turn ${idx}`,
              label: "neutral",
            },
            metadata: {
              generator_mode: "promise",
              index: idx,
            },
          };
        });
      },
    };
  },
});
