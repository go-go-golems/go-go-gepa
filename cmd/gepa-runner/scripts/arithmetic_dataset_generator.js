const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "examples.arithmetic",
  name: "Arithmetic Dataset Generator",
  registryIdentifier: "local.examples",
  create() {
    return {
      generateOne(input, options) {
        const rng = options && options.rng;
        const a = rng ? rng.intN(50) + 1 : 1;
        const b = rng ? rng.intN(50) + 1 : 1;
        return {
          row: {
            question: `${a} + ${b}`,
            answer: String(a + b),
          },
          metadata: {
            difficulty: "easy",
            index: input.index,
          },
        };
      },
    };
  },
});

