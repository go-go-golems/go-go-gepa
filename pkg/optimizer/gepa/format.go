package gepa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// DefaultReflectionPromptTemplate is adapted from GEPA's default "InstructionProposalSignature" prompt.
// It MUST contain "<curr_param>" and "<side_info>" placeholders.
const DefaultReflectionPromptTemplate = "" +
	"I provided an assistant with the following instructions to perform a task for me:\n" +
	"```\n" +
	"<curr_param>\n" +
	"```\n\n" +
	"The following are examples of different task inputs provided to the assistant along with the assistant's response for each of them, and some feedback on how the assistant's response could be better:\n" +
	"```\n" +
	"<side_info>\n" +
	"```\n\n" +
	"Your task is to write a new instruction for the assistant.\n\n" +
	"Read the inputs carefully and identify the input format and infer detailed task description about the task I wish to solve with the assistant.\n\n" +
	"Read all the assistant responses and the corresponding feedback. Identify all niche and domain specific factual information about the task and include it in the instruction, as a lot of it may not be available to the assistant in the future. The assistant may have utilized a generalizable strategy to solve the task, if so, include that in the instruction as well.\n\n" +
	"Provide the new instructions within ``` blocks."

// DefaultMergePromptTemplate asks an LLM to merge two candidate instructions.
// It MUST contain "<param_a>", "<param_b>", "<side_info_a>", and "<side_info_b>" placeholders.
const DefaultMergePromptTemplate = "" +
	"I have two different instructions for an assistant to perform the same task.\n" +
	"Instruction A:\n" +
	"```\n" +
	"<param_a>\n" +
	"```\n\n" +
	"Instruction B:\n" +
	"```\n" +
	"<param_b>\n" +
	"```\n\n" +
	"Below are evaluation examples (inputs, assistant outputs, scores, and feedback) for each instruction.\n\n" +
	"Examples for Instruction A:\n" +
	"```\n" +
	"<side_info_a>\n" +
	"```\n\n" +
	"Examples for Instruction B:\n" +
	"```\n" +
	"<side_info_b>\n" +
	"```\n\n" +
	"Your task is to write a single merged instruction that combines the best parts of A and B, fixes their failures, and stays concise and unambiguous.\n\n" +
	"Provide the merged instruction within ``` blocks."

// FormatSideInfo renders a minibatch of evaluations into a markdown-ish text block.
// This text becomes the "<side_info>" input in the reflection prompt template.
func FormatSideInfo(examples []any, evals []ExampleEval, maxChars int) string {
	var b bytes.Buffer
	for i, ev := range evals {
		ex := any(nil)
		if ev.ExampleIndex >= 0 && ev.ExampleIndex < len(examples) {
			ex = examples[ev.ExampleIndex]
		}

		fmt.Fprintf(&b, "### Example %d\n", i+1)

		if ex != nil {
			fmt.Fprintf(&b, "#### Input\n")
			writeAsPrettyJSON(&b, ex)
			b.WriteString("\n")
		}

		if ev.Result.Output != nil {
			fmt.Fprintf(&b, "#### Assistant Response\n")
			writeAsPrettyJSON(&b, ev.Result.Output)
			b.WriteString("\n")
		}

		fmt.Fprintf(&b, "#### Score\n%.6f\n\n", ev.Result.Score)

		if len(ev.Result.Objectives) > 0 {
			fmt.Fprintf(&b, "#### Objective Scores\n")
			writeAsPrettyJSON(&b, ev.Result.Objectives)
			b.WriteString("\n")
		}

		if ev.Result.Feedback != nil {
			fmt.Fprintf(&b, "#### Feedback\n")
			writeAsPrettyJSON(&b, ev.Result.Feedback)
			b.WriteString("\n")
		}

		if ev.Result.Trace != nil {
			fmt.Fprintf(&b, "#### Trace\n")
			writeAsPrettyJSON(&b, ev.Result.Trace)
			b.WriteString("\n")
		}

		b.WriteString("\n")
		if maxChars > 0 && b.Len() > maxChars {
			// Truncate gracefully at a line boundary.
			s := b.String()
			if len(s) > maxChars {
				s = s[:maxChars]
			}
			s = strings.TrimRight(s, "\n") + "\n\n[TRUNCATED]\n"
			return s
		}
	}
	return b.String()
}

// FormatSideInfoForKey is like FormatSideInfo, but (optionally) narrows Trace and Feedback
// to the sub-field keyed by paramKey when those values are JSON objects.
//
// This is useful for multi-parameter (multi-module) optimization, where evaluators may
// return per-module traces/feedback as an object keyed by the candidate parameter name.
func FormatSideInfoForKey(examples []any, evals []ExampleEval, paramKey string, maxChars int) string {
	paramKey = strings.TrimSpace(paramKey)
	if paramKey == "" {
		return FormatSideInfo(examples, evals, maxChars)
	}

	filtered := make([]ExampleEval, 0, len(evals))
	for _, ev := range evals {
		res := ev.Result
		if sub, ok := pickKeyedSubvalue(res.Trace, paramKey); ok {
			res.Trace = sub
		}
		if sub, ok := pickKeyedSubvalue(res.Feedback, paramKey); ok {
			res.Feedback = sub
		}
		filtered = append(filtered, ExampleEval{ExampleIndex: ev.ExampleIndex, Result: res})
	}

	return FormatSideInfo(examples, filtered, maxChars)
}

func pickKeyedSubvalue(v any, key string) (any, bool) {
	if v == nil {
		return nil, false
	}
	switch m := v.(type) {
	case map[string]any:
		if vv, ok := m[key]; ok {
			return vv, true
		}
	}
	return nil, false
}

func writeAsPrettyJSON(b *bytes.Buffer, v any) {
	blob, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Fprintf(b, "%v\n\n", v)
		return
	}
	b.Write(blob)
	b.WriteString("\n\n")
}
