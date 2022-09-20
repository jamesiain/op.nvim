package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/bmatcuk/doublestar/v4"
)

type LineDiagnostic struct {
	// number or null if workspace diagnostics
	BufNr *int `json:"bufnr"`
	// string or null if current buffer diagnostics
	File       *string `json:"file"`
	Line       int     `json:"line"`
	ColStart   int     `json:"col_start"`
	ColEnd     int     `json:"col_end"`
	SecretType string  `json:"secret_type"`
}

type LineDiagnosticRequest struct {
	// number or null if workspace diagnostics
	BufNr *int `json:"bufnr"`
	// string or null if current buffer diagnostics
	File   *string `json:"file"`
	LineNr int     `json:"linenr"`
	Text   string  `json:"text"`
}

// these types create too many false positives
var ignoredSecretTypes = []string{
	"username",
	"url",
}

func isIgnoredPattern(pattern FieldPattern) bool {
	for _, patternType := range ignoredSecretTypes {
		if pattern.FieldTitle == patternType {
			return true
		}
	}

	return false
}

func formatSecretType(pattern FieldPattern) string {
	if &pattern.ItemTitle != nil && len(pattern.ItemTitle) > 0 {
		return fmt.Sprintf("%s %s", pattern.ItemTitle, pattern.FieldTitle)
	}

	return pattern.FieldTitle
}

func lineMatches(pattern FieldPattern, line string) [][]int {
	if isIgnoredPattern(pattern) {
		return nil
	}

	return pattern.Pattern.FindAllStringIndex(line, -1)
}

func validLineRequests(lineRequests []LineDiagnosticRequest) []LineDiagnosticRequest {
	validRequests := make([]LineDiagnosticRequest, len(lineRequests))
	for _, req := range lineRequests {
		if &req.Text != nil && len(req.Text) > 0 {
			validRequests = append(validRequests, req)
		}
	}

	return validRequests
}

func generateDiagnostics(req LineDiagnosticRequest) []LineDiagnostic {
	diagnostics := []LineDiagnostic{}
	linenr := req.LineNr
	line := req.Text
	if &line == nil || len(line) == 0 {
		return diagnostics
	}

	for _, pattern := range FIELD_PATTERNS {
		secretType := formatSecretType(pattern)
		for _, match := range lineMatches(pattern, line) {
			diagnostics = append(diagnostics, LineDiagnostic{
				BufNr:      req.BufNr,
				File:       req.File,
				Line:       linenr,
				ColStart:   match[0],
				ColEnd:     match[1],
				SecretType: secretType,
			})
		}
	}

	return diagnostics
}

func analyzeBuffer(lineRequests []LineDiagnosticRequest) []LineDiagnostic {
	results := []LineDiagnostic{}
	for _, req := range lineRequests {
		results = append(results, generateDiagnostics(req)...)
	}

	return results
}

func analyzeBufferJson(requestId string, lineRequests []LineDiagnosticRequest) {
	results := analyzeBuffer(lineRequests)
	result, err := json.Marshal(results)

	if err != nil {
		Async.Err(requestId, err)
	} else {
		json := string(result)
		Async.Success(requestId, json)
	}
}

func collectWorkspaceFiles(globs []string) ([]string, error) {
	files := []string{}
	for _, glob := range globs {
		globFiles, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}

		files = append(files, globFiles...)
	}

	return files, nil
}

func getDiagnosticsForFile(filepath string, diagnostics *[]LineDiagnostic, wg *sync.WaitGroup) {
	wg.Add(1)
	diagnosticRequests := []LineDiagnosticRequest{}
	file, openErr := os.Open(filepath)
	if openErr != nil {
		// fail gracefully
		file.Close()
		wg.Done()
		return
	}
	scanner := bufio.NewScanner(file)
	linenr := 0
	for scanner.Scan() {
		req := LineDiagnosticRequest{
			File:   &filepath,
			BufNr:  nil,
			LineNr: linenr,
			Text:   scanner.Text(),
		}
		diagnosticRequests = append(diagnosticRequests, req)
		linenr += 1
	}

	file.Close()
	*diagnostics = append(*diagnostics, analyzeBuffer(diagnosticRequests)...)
	wg.Done()
}

func isIgnoredPath(path string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		match, _ := doublestar.PathMatch(pattern, path)
		if match {
			return true
		}
	}

	return false
}

func genDiagnosticsForWorkspace(ignorePatterns []string) []LineDiagnostic {
	diagnostics := []LineDiagnostic{}
	wg := sync.WaitGroup{}

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() || isIgnoredPath(path, ignorePatterns) {
			return nil
		}

		go getDiagnosticsForFile(path, &diagnostics, &wg)
		return nil
	})

	wg.Wait()
	return diagnostics
}

func genDiagnosticRequestsForWorkspaceJson(requestId string, ignorePatterns []string) {
	diagnostics := genDiagnosticsForWorkspace(ignorePatterns)
	json, err := json.Marshal(diagnostics)
	if err != nil {
		Async.Err(requestId, err)
	} else {
		jsonStr := string(json)
		Async.Success(requestId, jsonStr)
	}
}

func OpAnalyzeBufferAsync(args []string) error {
	if len(args) != 2 {
		return errors.New("Need exactly 2 arguments (request ID, then buffer line requests)")
	}

	var lineRequests []LineDiagnosticRequest
	jsonParseErr := json.Unmarshal([]byte(args[1]), &lineRequests)
	if jsonParseErr != nil {
		return jsonParseErr
	}

	go analyzeBufferJson(args[0], lineRequests)

	return nil
}

func OpAnalyzeWorkspaceAsync(args []string) error {
	if len(args) < 2 {
		return errors.New("Need at least 2 arguments (request ID, then globbing patterns)")
	}

	requestId := args[0]
	ignorePatterns := args[1:]

	go genDiagnosticRequestsForWorkspaceJson(requestId, ignorePatterns)

	return nil
}
