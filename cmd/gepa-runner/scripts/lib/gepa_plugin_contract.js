const OPTIMIZER_PLUGIN_API_VERSION = "gepa.optimizer/v1";

function toTrimmedString(v) {
  return String(v == null ? "" : v).trim();
}

function defineOptimizerPlugin(descriptor) {
  const d = (descriptor && typeof descriptor === "object") ? descriptor : null;
  if (!d) {
    throw new Error("plugin descriptor must be an object");
  }

  const apiVersion = toTrimmedString(d.apiVersion || OPTIMIZER_PLUGIN_API_VERSION);
  if (apiVersion !== OPTIMIZER_PLUGIN_API_VERSION) {
    throw new Error(`unsupported plugin descriptor apiVersion "${apiVersion}" (expected "${OPTIMIZER_PLUGIN_API_VERSION}")`);
  }

  const kind = toTrimmedString(d.kind || "optimizer");
  if (kind !== "optimizer") {
    throw new Error(`plugin descriptor kind must be "optimizer", got "${kind}"`);
  }

  const id = toTrimmedString(d.id);
  if (!id) {
    throw new Error("plugin descriptor id is required");
  }

  const name = toTrimmedString(d.name);
  if (!name) {
    throw new Error("plugin descriptor name is required");
  }

  if (typeof d.create !== "function") {
    throw new Error("plugin descriptor create must be a function");
  }

  return Object.freeze({
    apiVersion,
    kind,
    id,
    name,
    create: d.create,
  });
}

module.exports = {
  OPTIMIZER_PLUGIN_API_VERSION,
  defineOptimizerPlugin,
};
