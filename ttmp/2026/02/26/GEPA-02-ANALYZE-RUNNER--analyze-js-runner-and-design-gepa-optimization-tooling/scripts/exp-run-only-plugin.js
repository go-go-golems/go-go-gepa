module.exports = {
  apiVersion: "gepa.optimizer/v1",
  kind: "optimizer",
  id: "experiment.run_only",
  name: "Experiment: Run Only Plugin",
  create() {
    return {
      dataset() {
        return [{ question: "2+2", answer: "4" }];
      },
      run(input) {
        return {
          output: {
            text: String((input && input.example && input.example.answer) || "")
          }
        };
      }
    };
  }
};
