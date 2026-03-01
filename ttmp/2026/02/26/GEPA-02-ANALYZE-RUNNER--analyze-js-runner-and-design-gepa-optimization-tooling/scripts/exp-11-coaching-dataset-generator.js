const { defineDatasetGenerator, DATASET_GENERATOR_API_VERSION } = require("gepa/plugins");
const gp = require("geppetto");

function toTrimmedString(v) {
  return String(v == null ? "" : v).trim();
}

function resolveAssistantText(out) {
  const blocks = (out && Array.isArray(out.blocks)) ? out.blocks : [];
  return blocks
    .filter((b) => b && b.payload && typeof b.payload.text === "string")
    .filter((b) => b.kind === gp.consts.BlockKind.LLM_TEXT || b.kind === "assistant")
    .map((b) => toTrimmedString(b.payload.text))
    .join("\n")
    .trim();
}

function normalizeEngineOptions(options) {
  const opts = (options && typeof options === "object") ? options : {};
  const profile = toTrimmedString(opts.profile || "");
  const engineOptions = (opts.engineOptions && typeof opts.engineOptions === "object") ? opts.engineOptions : null;
  return { profile, engineOptions };
}

function createEngine(options) {
  const resolved = normalizeEngineOptions(options);
  if (resolved.engineOptions && Object.keys(resolved.engineOptions).length > 0) {
    return gp.engines.fromConfig(resolved.engineOptions);
  }
  return gp.engines.fromProfile(resolved.profile || "", {});
}

function boolFromAny(v, fallback) {
  if (typeof v === "boolean") {
    return v;
  }
  const s = toTrimmedString(v).toLowerCase();
  if (!s) {
    return fallback;
  }
  if (s === "1" || s === "true" || s === "yes" || s === "y" || s === "on") {
    return true;
  }
  if (s === "0" || s === "false" || s === "no" || s === "n" || s === "off") {
    return false;
  }
  return fallback;
}

function resolveStopReason(out) {
  const metadata = (out && out.metadata && typeof out.metadata === "object") ? out.metadata : {};
  const stopReasonKey = (gp.consts && gp.consts.TurnMetadataKeys && gp.consts.TurnMetadataKeys.STOP_REASON)
    ? gp.consts.TurnMetadataKeys.STOP_REASON
    : "stop_reason";
  return toTrimmedString(metadata[stopReasonKey] || metadata.stop_reason || metadata.stopReason || "");
}

function isTokenLimitStopReason(stopReason) {
  const s = toTrimmedString(stopReason).toLowerCase();
  if (!s) {
    return false;
  }
  return (
    s.includes("max_tokens") ||
    s.includes("max_output_tokens") ||
    s.includes("token_limit") ||
    s.includes("output_token") ||
    s.includes("length") ||
    s.includes("truncat")
  );
}

function appendWithOverlap(base, addition) {
  const left = String(base || "");
  const right = String(addition || "");
  if (!left) {
    return right;
  }
  if (!right) {
    return left;
  }

  const maxOverlap = Math.min(left.length, right.length, 4096);
  for (let size = maxOverlap; size > 0; size--) {
    if (left.slice(left.length - size) === right.slice(0, size)) {
      return left + right.slice(size);
    }
  }
  return left + right;
}

function buildContinuationPrompt() {
  return [
    "Continue your previous response from exactly where it ended.",
    "Output only the remaining JSON text.",
    "Do not repeat any text already produced.",
    "Do not add markdown fences or commentary.",
  ].join("\n");
}

function runPrompt(prompt, options) {
  const engine = createEngine(options);
  const session = gp.createBuilder().withEngine(engine).buildSession();
  const maxContinuationAttempts = intFromAny(options && options.maxContinuationAttempts, 4);
  const streamResponses = boolFromAny(options && options.streamResponses, false);
  const systemPrompt = toTrimmedString(options && options.systemPrompt);
  const runTags = (options && options.tags && typeof options.tags === "object") ? options.tags : {};
  if (systemPrompt) {
    session.append(gp.turns.newTurn({
      blocks: [gp.turns.newSystemBlock(systemPrompt)],
    }));
  }

  let accumulatedText = "";
  let lastStopReason = "";
  let lastParseError = null;

  for (let attempt = 0; attempt <= maxContinuationAttempts; attempt++) {
    const promptText = (attempt === 0)
      ? String(prompt || "")
      : buildContinuationPrompt();

    const turn = gp.turns.newTurn({
      blocks: [gp.turns.newUserBlock(promptText)],
    });
    const out = session.run(turn, { tags: runTags });
    const chunk = resolveAssistantText(out);
    lastStopReason = resolveStopReason(out);
    accumulatedText = appendWithOverlap(accumulatedText, chunk);

    if (streamResponses) {
      const preview = toTrimmedString(chunk).slice(0, 160);
      console.error(`[exp-11 stream] attempt=${attempt + 1} chunk_chars=${chunk.length} stop_reason=${lastStopReason || "none"} preview=${JSON.stringify(preview)}`);
    }

    try {
      const parsed = extractJSON(accumulatedText);
      return {
        text: accumulatedText,
        parsed,
        stopReason: lastStopReason,
        attempts: attempt + 1,
      };
    } catch (err) {
      lastParseError = err;
    }

    if (!isTokenLimitStopReason(lastStopReason)) {
      break;
    }
  }

  const parseErrText = lastParseError && lastParseError.message ? lastParseError.message : "invalid json output";
  const reasonSuffix = lastStopReason ? ` (stop_reason=${lastStopReason})` : "";
  throw new Error(`unable to assemble valid JSON from LLM output after continuation attempts: ${parseErrText}${reasonSuffix}`);
}

function extractJSON(text) {
  const raw = toTrimmedString(text);
  if (!raw) {
    throw new Error("empty LLM output");
  }

  const fenced = raw.match(/```(?:json)?\s*([\s\S]*?)\s*```/i);
  const candidate = fenced && fenced[1] ? fenced[1].trim() : raw;

  try {
    return JSON.parse(candidate);
  } catch (_err) {
    const start = candidate.indexOf("{");
    const end = candidate.lastIndexOf("}");
    if (start >= 0 && end > start) {
      return JSON.parse(candidate.slice(start, end + 1));
    }
    throw _err;
  }
}

function renderTemplate(template, variables) {
  const tpl = String(template || "");
  return tpl.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g, (_m, key) => {
    const value = variables[key];
    return value == null ? "" : String(value);
  });
}

function intFromAny(v, fallback) {
  if (typeof v === "number" && Number.isFinite(v)) {
    return Math.max(0, Math.floor(v));
  }
  const n = Number(String(v == null ? "" : v).trim());
  if (!Number.isFinite(n)) {
    return fallback;
  }
  return Math.max(0, Math.floor(n));
}

module.exports = defineDatasetGenerator({
  apiVersion: DATASET_GENERATOR_API_VERSION,
  kind: "dataset-generator",
  id: "examples.coaching-entity-sentiment",
  name: "Coaching Entity/Sentiment Longitudinal Generator",
  registryIdentifier: "local.examples",

  create() {
    return {
      generateOne(input, options) {
        const inObj = (input && typeof input === "object") ? input : {};
        const vars = (inObj.variables && typeof inObj.variables === "object") ? inObj.variables : {};
        const cfg = (options && typeof options === "object" && options.config && typeof options.config === "object") ? options.config : {};
        const rng = options && options.rng;

        const idx = intFromAny(inObj.index, 0);
        const sessionsMin = intFromAny(vars.sessions_min, 5);
        const sessionsMax = Math.max(sessionsMin, intFromAny(vars.sessions_max, 6));
        const spanMin = intFromAny(vars.time_span_months_min, 2);
        const spanMax = Math.max(spanMin, intFromAny(vars.time_span_months_max, 4));
        const complexity = toTrimmedString(vars.complexity || "medium");
        const casePrefix = toTrimmedString(vars.case_prefix || "COACH_CASE");

        const sessionCount = rng ? (sessionsMin + rng.intN((sessionsMax - sessionsMin) + 1)) : sessionsMin;
        const timeSpanMonths = rng ? (spanMin + rng.intN((spanMax - spanMin) + 1)) : spanMin;
        const caseId = `${casePrefix}_${String(idx + 1).padStart(3, "0")}`;

        const baseTemplate = (inObj.promptSpec && typeof inObj.promptSpec === "object" && typeof inObj.promptSpec.user_template === "string")
          ? inObj.promptSpec.user_template
          : "";

        const generationPrompt = renderTemplate(baseTemplate, {
          case_id: caseId,
          session_count: sessionCount,
          time_span_months: timeSpanMonths,
          complexity: complexity,
        });

        const fallbackPrompt = [
          "Generate one synthetic longitudinal coaching case for NLP benchmarking.",
          `Case ID: ${caseId}`,
          `Sessions: ${sessionCount}`,
          `Time span months: ${timeSpanMonths}`,
          `Complexity: ${complexity}`,
          "Use realistic spoken coaching/therapy transcript style with disfluencies and non-linear emotional progression.",
          "Surface at least 6 named people and evolving relationships.",
          "Return strict JSON with keys: case_id, transcript, ground_truth.",
          "transcript: array of sessions {session_id, session_date, turns:[{speaker,text}]}",
          "ground_truth must include: entities, mentions_by_session, relationships, evolution_timeline, client_trajectory.",
          "No markdown, no prose outside JSON.",
        ].join("\n");

        const prompt = generationPrompt || fallbackPrompt;
        const promptResult = runPrompt(prompt, {
          profile: options && options.profile,
          engineOptions: options && options.engineOptions,
          tags: options && options.tags,
          systemPrompt: inObj.promptSpec && inObj.promptSpec.system,
          maxContinuationAttempts: intFromAny(vars.max_continuation_attempts, 4),
          streamResponses: boolFromAny(vars.stream_responses, false),
        });
        const parsed = promptResult.parsed;

        const transcript = Array.isArray(parsed.transcript) ? parsed.transcript : [];
        const groundTruth = (parsed.ground_truth && typeof parsed.ground_truth === "object") ? parsed.ground_truth : {};

        if (!parsed.case_id) {
          parsed.case_id = caseId;
        }
        if (!Array.isArray(transcript) || transcript.length === 0) {
          throw new Error("generated case missing transcript sessions");
        }
        if (!groundTruth || typeof groundTruth !== "object" || Object.keys(groundTruth).length === 0) {
          throw new Error("generated case missing ground_truth");
        }

        const row = {
          case_id: String(parsed.case_id),
          transcript: transcript,
          ground_truth: groundTruth,
          generation_context: {
            session_count: sessionCount,
            time_span_months: timeSpanMonths,
            complexity: complexity,
            config_name: toTrimmedString(cfg.name || ""),
          },
        };

        return {
          row,
          metadata: {
            case_id: row.case_id,
            plugin_id: "examples.coaching-entity-sentiment",
            sessions_generated: transcript.length,
            profile: toTrimmedString(options && options.profile),
            llm_attempts: promptResult.attempts,
            llm_stop_reason: toTrimmedString(promptResult.stopReason),
            llm_used_continuation: promptResult.attempts > 1,
          },
        };
      },
    };
  },
});
