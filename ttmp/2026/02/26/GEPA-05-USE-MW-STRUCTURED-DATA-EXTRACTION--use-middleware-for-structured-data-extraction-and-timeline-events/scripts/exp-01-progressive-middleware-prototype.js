#!/usr/bin/env node

const fs = require("fs");
const path = require("path");

function parseArgs(argv) {
  const out = {};
  for (let i = 2; i < argv.length; i += 1) {
    const token = argv[i];
    if (!token.startsWith("--")) {
      continue;
    }
    const key = token.slice(2);
    const value = argv[i + 1];
    if (value && !value.startsWith("--")) {
      out[key] = value;
      i += 1;
    } else {
      out[key] = "true";
    }
  }
  return out;
}

function readFirstJSONLRow(filePath) {
  const raw = fs.readFileSync(filePath, "utf8");
  const line = raw
    .split("\n")
    .map((x) => x.trim())
    .find((x) => x.length > 0);
  if (!line) {
    throw new Error(`no JSONL rows found in ${filePath}`);
  }
  return JSON.parse(line);
}

function toArray(v) {
  return Array.isArray(v) ? v : [];
}

function toObject(v) {
  return v && typeof v === "object" ? v : {};
}

function toString(v) {
  return String(v == null ? "" : v).trim();
}

function summarizeSession(session) {
  const obj = toObject(session);
  const turns = toArray(obj.turns);
  let coachTurns = 0;
  let clientTurns = 0;
  const highlights = [];

  for (let i = 0; i < turns.length; i += 1) {
    const t = toObject(turns[i]);
    const speaker = toString(t.speaker).toLowerCase();
    const text = toString(t.text);
    if (speaker === "coach") {
      coachTurns += 1;
    } else if (speaker === "client") {
      clientTurns += 1;
    }
    if (highlights.length < 2 && text) {
      highlights.push(text.slice(0, 160));
    }
  }

  return {
    session_id: obj.session_id,
    session_date: obj.session_date,
    coach_turns: coachTurns,
    client_turns: clientTurns,
    turn_count: turns.length,
    highlights,
  };
}

function createEmitter(seed) {
  const base = toObject(seed);
  const correlation = {
    message_id: toString(base.message_id || "msg-gepa-05-prototype-001"),
    session_id: toString(base.session_id || "sess-gepa-05-prototype"),
    inference_id: toString(base.inference_id || "inf-gepa-05-prototype"),
    turn_id: toString(base.turn_id || "turn-gepa-05-prototype"),
  };

  let seq = 0;
  const events = [];

  function emit(eventType, stage, payload) {
    seq += 1;
    events.push({
      event_id: `${correlation.message_id}:${String(seq).padStart(4, "0")}`,
      event_type: eventType,
      stage,
      timestamp: new Date().toISOString(),
      correlation,
      payload: toObject(payload),
    });
  }

  return { emit, events, correlation };
}

function stageDescriptor(packageName, typeName, version, instruction) {
  return {
    package: packageName,
    type: typeName,
    version,
    open_tag: `<${packageName}:${typeName}:${version}>`,
    close_tag: `</${packageName}:${typeName}:${version}>`,
    instruction,
  };
}

function emitStagePrompt(emit, stage, descriptor, extra) {
  emit("middleware.prompt.generated", stage, {
    descriptor,
    prompt_contract: {
      required_format: `${descriptor.open_tag} + fenced YAML + ${descriptor.close_tag}`,
      instruction: descriptor.instruction,
    },
    context: toObject(extra),
  });
}

function entitiesMiddleware(state) {
  const stage = "entities";
  const descriptor = stageDescriptor(
    "gepa",
    "entities",
    "v1",
    "Extract canonical people/entities and stable IDs before any relationship extraction.",
  );
  emitStagePrompt(state.emit, stage, descriptor, { case_id: state.caseId });
  state.emit("structured.block.started", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:entities`,
  });

  const entities = toArray(toObject(state.row.ground_truth).entities);
  const entityIndex = {};
  for (let i = 0; i < entities.length; i += 1) {
    const e = toObject(entities[i]);
    const entityId = toString(e.entity_id || `E${String(i + 1).padStart(3, "0")}`);
    entityIndex[entityId] = e;
    state.emit("entity.upsert", stage, {
      entity_id: entityId,
      canonical_name: toString(e.canonical_name),
      entity_type: toString(e.entity_type),
      first_mentioned_session: e.first_mentioned_session,
      aliases: toArray(e.aliases),
    });
  }

  state.emit("structured.block.completed", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:entities`,
    entity_count: entities.length,
    success: true,
  });

  state.entities = entityIndex;
}

function relationshipsMiddleware(state) {
  const stage = "relationships";
  const descriptor = stageDescriptor(
    "gepa",
    "relationships",
    "v1",
    "Given known entities, extract directed relationships with IDs and first mention session.",
  );
  emitStagePrompt(state.emit, stage, descriptor, {
    known_entity_ids: Object.keys(toObject(state.entities)),
  });
  state.emit("structured.block.started", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:relationships`,
  });

  const relationships = toArray(toObject(state.row.ground_truth).relationships);
  for (let i = 0; i < relationships.length; i += 1) {
    const rel = toObject(relationships[i]);
    state.emit("relationship.upsert", stage, {
      relationship_id: toString(rel.relationship_id || `R${String(i + 1).padStart(3, "0")}`),
      source_entity: toString(rel.source_entity),
      target_entity: toString(rel.target_entity),
      relationship_type: toString(rel.relationship_type),
      relationship_label: toString(rel.relationship_label),
      first_mentioned_session: rel.first_mentioned_session,
    });
  }

  state.emit("structured.block.completed", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:relationships`,
    relationship_count: relationships.length,
    success: true,
  });
}

function summariesMiddleware(state) {
  const stage = "discussion_summaries";
  const descriptor = stageDescriptor(
    "gepa",
    "discussion-summary",
    "v1",
    "As each session is processed, append a concise summary and keep a running longitudinal summary.",
  );
  emitStagePrompt(state.emit, stage, descriptor, {
    transcript_sessions: toArray(state.row.transcript).length,
  });
  state.emit("structured.block.started", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:discussion-summary`,
  });

  const transcript = toArray(state.row.transcript);
  const summaries = [];
  for (let i = 0; i < transcript.length; i += 1) {
    const item = summarizeSession(transcript[i]);
    summaries.push(item);
    const runningSummary = summaries
      .map((x) => `S${x.session_id}: ${x.highlights.join(" | ")}`)
      .join(" || ");

    state.emit("discussion.summary.delta", stage, {
      session_id: item.session_id,
      session_date: item.session_date,
      turn_count: item.turn_count,
      highlights: item.highlights,
      running_summary: runningSummary,
    });
  }

  state.summaries = summaries;
  state.emit("structured.block.completed", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:discussion-summary`,
    summary_count: summaries.length,
    success: true,
  });
}

function timelineMiddleware(state) {
  const stage = "timeline";
  const descriptor = stageDescriptor(
    "gepa",
    "timeline-events",
    "v1",
    "Emit timeline-ready events keyed by relationship_id + session_id, enriched with sentiment and summary context.",
  );
  emitStagePrompt(state.emit, stage, descriptor, {
    relationship_count: toArray(toObject(state.row.ground_truth).relationships).length,
  });
  state.emit("structured.block.started", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:timeline`,
  });

  const points = toArray(toObject(state.row.ground_truth).evolution_timeline);
  for (let i = 0; i < points.length; i += 1) {
    const p = toObject(points[i]);
    const timelineId = `${toString(p.relationship_id)}:s${toString(p.session_id)}`;
    state.emit("timeline.event.upsert", stage, {
      timeline_event_id: timelineId,
      relationship_id: toString(p.relationship_id),
      session_id: p.session_id,
      session_date: p.session_date,
      emotional_valence: p.emotional_valence,
      key_emotions_expressed: toArray(p.key_emotions_expressed),
      client_stance: toString(p.client_stance),
      summary: toString(p.summary),
      supporting_turn_indices: toArray(p.supporting_turn_indices),
    });
  }

  state.emit("structured.block.completed", stage, {
    descriptor,
    item_id: `${state.correlation.message_id}:timeline`,
    timeline_event_count: points.length,
    success: true,
  });
}

function main() {
  const args = parseArgs(process.argv);
  const cwd = process.cwd();
  const defaultInput = path.resolve(
    cwd,
    "go-go-gepa/ttmp/2026/02/26/GEPA-02-ANALYZE-RUNNER--analyze-js-runner-and-design-gepa-optimization-tooling/scripts/exp-11-out/coaching-entity-sentiment-small.jsonl",
  );
  const inputPath = path.resolve(cwd, args.input || defaultInput);
  const outPath = path.resolve(cwd, args.out || "exp-01-events.jsonl");
  const summaryPath = path.resolve(cwd, args.summary || "exp-01-summary.json");

  const row = readFirstJSONLRow(inputPath);
  const caseId = toString(row.case_id || "unknown-case");
  const emitter = createEmitter({ message_id: `msg-${caseId.toLowerCase()}` });

  const state = {
    row,
    caseId,
    emit: emitter.emit,
    correlation: emitter.correlation,
    entities: {},
    summaries: [],
  };

  state.emit("pipeline.started", "pipeline", {
    case_id: caseId,
    source_jsonl: inputPath,
    stage_order: ["entities", "relationships", "discussion_summaries", "timeline"],
  });

  entitiesMiddleware(state);
  relationshipsMiddleware(state);
  summariesMiddleware(state);
  timelineMiddleware(state);

  state.emit("pipeline.completed", "pipeline", {
    case_id: caseId,
    total_events_emitted: emitter.events.length + 1,
  });

  fs.mkdirSync(path.dirname(outPath), { recursive: true });
  const jsonl = emitter.events.map((e) => JSON.stringify(e)).join("\n") + "\n";
  fs.writeFileSync(outPath, jsonl);

  const summary = {
    case_id: caseId,
    input_path: inputPath,
    output_events_path: outPath,
    event_count: emitter.events.length,
    stage_counts: emitter.events.reduce((acc, ev) => {
      acc[ev.stage] = (acc[ev.stage] || 0) + 1;
      return acc;
    }, {}),
    first_event: emitter.events[0] || null,
    last_event: emitter.events[emitter.events.length - 1] || null,
  };
  fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2) + "\n");

  console.log(`[exp-01] input=${inputPath}`);
  console.log(`[exp-01] wrote events=${outPath} count=${emitter.events.length}`);
  console.log(`[exp-01] wrote summary=${summaryPath}`);
}

main();
