package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gepaopt "github.com/go-go-golems/go-go-gepa/pkg/optimizer/gepa"
	"gopkg.in/yaml.v3"
)

func resolveSeedText(seed, seedFile string) (string, error) {
	if strings.TrimSpace(seed) != "" {
		return seed, nil
	}
	if strings.TrimSpace(seedFile) == "" {
		return "", nil
	}
	blob, err := os.ReadFile(seedFile)
	if err != nil {
		return "", err
	}
	return string(blob), nil
}

func loadSeedCandidateFile(path string) (gepaopt.Candidate, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("seed candidate path is empty")
	}
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	var v any
	var parseErr error
	switch ext {
	case ".yaml", ".yml":
		parseErr = yaml.Unmarshal(blob, &v)
	default:
		parseErr = json.Unmarshal(blob, &v)
	}
	if parseErr != nil {
		var v2 any
		if ext == ".yaml" || ext == ".yml" {
			if err2 := json.Unmarshal(blob, &v2); err2 == nil {
				v = v2
				parseErr = nil
			}
		} else {
			if err2 := yaml.Unmarshal(blob, &v2); err2 == nil {
				v = v2
				parseErr = nil
			}
		}
	}
	if parseErr != nil {
		return nil, parseErr
	}

	m, ok := v.(map[string]any)
	if !ok {
		if m3, ok3 := v.(map[interface{}]interface{}); ok3 {
			m = map[string]any{}
			for k, vv := range m3 {
				m[fmt.Sprintf("%v", k)] = vv
			}
		} else {
			return nil, fmt.Errorf("seed candidate must be an object/map, got %T", v)
		}
	}

	cand := gepaopt.Candidate{}
	for k, vv := range m {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		cand[k] = coerceToString(vv)
	}
	return cand, nil
}

func coerceToString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int16:
		return strconv.FormatInt(int64(t), 10)
	case int8:
		return strconv.FormatInt(int64(t), 10)
	case uint:
		return strconv.FormatUint(uint64(t), 10)
	case uint64:
		return strconv.FormatUint(t, 10)
	case uint32:
		return strconv.FormatUint(uint64(t), 10)
	case uint16:
		return strconv.FormatUint(uint64(t), 10)
	case uint8:
		return strconv.FormatUint(uint64(t), 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(t), 'f', -1, 32)
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		if blob, err := json.Marshal(t); err == nil {
			return string(blob)
		}
		return fmt.Sprintf("%v", t)
	}
}

func loadDataset(path string) ([]any, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("dataset path is empty")
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jsonl":
		return loadJSONL(path)
	case ".json":
		return loadJSONArray(path)
	default:
		// Try jsonl first, then json.
		if xs, err := loadJSONL(path); err == nil && len(xs) > 0 {
			return xs, nil
		}
		return loadJSONArray(path)
	}
}

func loadJSONL(path string) ([]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var out []any
	sc := bufio.NewScanner(f)
	// Allow fairly long lines.
	buf := make([]byte, 0, 1024*1024)
	sc.Buffer(buf, 10*1024*1024)

	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var v any
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			return nil, fmt.Errorf("jsonl parse error at line %d: %w", lineNo, err)
		}
		out = append(out, v)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func loadJSONArray(path string) ([]any, error) {
	blob, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var v any
	if err := json.Unmarshal(blob, &v); err != nil {
		return nil, err
	}
	arr, ok := v.([]any)
	if ok {
		return arr, nil
	}
	return nil, fmt.Errorf("json dataset must be an array, got %T", v)
}
