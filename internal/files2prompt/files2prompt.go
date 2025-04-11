// Package files2prompt provides functionality for crawling directories,
// filtering files, and preparing content suitable for AI prompt ingestion.
package files2prompt

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	log "github.com/sirupsen/logrus"
	"github.com/toozej/files2prompt/pkg/config"
)

func readGitignore(path string) []string {
	gitignorePath := filepath.Join(path, ".gitignore")
	content, err := os.ReadFile(gitignorePath) // #nosec G304
	if err != nil {
		return []string{}
	}

	var rules []string
	for _, line := range strings.Split(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") {
			rules = append(rules, line)
		}
	}
	return rules
}

func shouldIgnore(path string, gitignoreRules []string) bool {
	base := filepath.Base(path)
	for _, rule := range gitignoreRules {
		matchedBase, _ := doublestar.Match(rule, base)
		if matchedBase {
			return true
		}
		matchedSlash, _ := doublestar.Match(rule, base+"/")
		if strings.HasSuffix(rule, "/") && matchedSlash {
			return true
		}
	}
	return false
}

func processPath(path string, config config.Config, writer io.Writer, gitignoreRules []string, globalIndex *int) error {
	// Handle current directory case
	if path == "." {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %v", err)
		}
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return processFile(path, config, writer, globalIndex)
	}

	return filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden files/directories unless specified
		if !config.IncludeHidden && strings.HasPrefix(filepath.Base(filePath), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply gitignore rules
		if config.IgnoreGitignore {
			if info.IsDir() {
				newRules := readGitignore(filePath)
				gitignoreRules = append(gitignoreRules, newRules...)
			}
			if shouldIgnore(filePath, gitignoreRules) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Apply ignore patterns to both files and directories
		if len(config.IgnorePatterns) > 0 {
			relPath, err := filepath.Rel(path, filePath)
			if err != nil {
				log.Warnf("Warning: Could not get relative path for %s: %v", filePath, err)
				relPath = filePath
			}

			for _, pattern := range config.IgnorePatterns {
				// Split pattern into individual paths if comma-separated
				for _, subPattern := range strings.Split(pattern, ",") {
					subPattern = strings.TrimSpace(subPattern)
					if subPattern == "" {
						continue
					}

					// Match against both the base name and the relative path
					baseMatch, _ := doublestar.Match(subPattern, filepath.Base(filePath))
					pathMatch, _ := doublestar.Match(subPattern, relPath)

					if baseMatch || pathMatch {
						if info.IsDir() {
							return filepath.SkipDir
						}
						return nil
					}

					// Handle directory-specific patterns
					if strings.HasSuffix(subPattern, "/") && info.IsDir() {
						dirPattern := strings.TrimSuffix(subPattern, "/")
						if match, _ := doublestar.Match(dirPattern, filepath.Base(filePath)); match {
							return filepath.SkipDir
						}
					}
				}
			}
		}

		// Apply extension filter only to files
		if len(config.Extensions) > 0 && !info.IsDir() {
			ext := filepath.Ext(filePath)
			var match bool
			for _, allowedExt := range config.Extensions {
				if ext == allowedExt {
					match = true
					break
				}
			}
			if !match {
				return nil
			}
		}

		if !info.IsDir() {
			return processFile(filePath, config, writer, globalIndex)
		}
		return nil
	})
}

func processFile(filePath string, config config.Config, writer io.Writer, globalIndex *int) error {
	content, err := os.ReadFile(filePath) // #nosec G304
	if err != nil {
		log.Warnf("Warning: Skipping file %s due to error: %v", filePath, err)
		return nil
	}

	lines := strings.Split(string(content), "\n")
	var processedContent strings.Builder

	// Process content with line numbers if enabled
	if config.LineNumbers {
		// Calculate padding for line numbers based on total lines
		padding := len(fmt.Sprintf("%d", len(lines)))
		format := fmt.Sprintf("%%%dd │ %%s\n", padding)

		for i, line := range lines {
			processedContent.WriteString(fmt.Sprintf(format, i+1, line))
		}
	} else {
		processedContent.WriteString(string(content))
	}

	if config.ClaudeXML {
		xmlOutput := fmt.Sprintf("<document index=\"%d\">\n<source>%s</source>\n<document_content>\n%s\n</document_content>\n</document>\n",
			*globalIndex, filePath, processedContent.String())
		*globalIndex++
		_, err = writer.Write([]byte(xmlOutput))
	} else {
		output := fmt.Sprintf("%s\n---\n%s\n---\n\n", filePath, processedContent.String())
		_, err = writer.Write([]byte(output))
	}

	return err
}

// Run executes the files2prompt logic using the provided config.
// It walks through each path, reads applicable files, and writes output
// either to stdout or a file depending on config.
func Run(config config.Config) error {
	log.Debugf("files2prompt pkg Run config config struct contains: %v\n", config)

	var writer io.Writer = os.Stdout
	var file *os.File

	if config.OutputFile != "" {
		var err error
		file, err = os.Create(config.OutputFile)
		if err != nil {
			log.Fatalf("Failed to create output file: %v", err)
		}
		defer file.Close()
		writer = file
	}

	globalIndex := 1
	var gitignoreRules []string

	if config.IgnoreGitignore {
		log.Debug("files2prompt pkg Run inside config.IgnoreGitignore check")
		for _, path := range config.Paths {
			gitignoreRules = append(gitignoreRules, readGitignore(filepath.Dir(path))...)
		}
	}

	if config.ClaudeXML {
		_, _ = writer.Write([]byte("<documents>\n"))
	}

	for _, path := range config.Paths {
		if err := processPath(path, config, writer, gitignoreRules, &globalIndex); err != nil {
			log.Errorf("Error processing path %s: %v", path, err)
		}
	}

	if config.ClaudeXML {
		_, _ = writer.Write([]byte("</documents>\n"))
	}
	return nil
}
