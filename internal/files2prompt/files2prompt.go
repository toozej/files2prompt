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
	content, err := os.ReadFile(gitignorePath)
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
		if !config.IgnoreGitignore {
			if info.IsDir() {
				dirGitignoreRules := append(gitignoreRules, readGitignore(filePath)...)
				if shouldIgnore(filePath, dirGitignoreRules) {
					return filepath.SkipDir
				}
			} else {
				if shouldIgnore(filePath, gitignoreRules) {
					return nil
				}
			}
		}

		// Apply extension filter
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

		// Apply ignore patterns
		if len(config.IgnorePatterns) > 0 && !info.IsDir() {
			base := filepath.Base(filePath)
			for _, pattern := range config.IgnorePatterns {
				if match, _ := doublestar.Match(pattern, base); match {
					return nil
				}
			}
		}

		if !info.IsDir() {
			return processFile(filePath, config, writer, globalIndex)
		}
		return nil
	})
}

func processFile(filePath string, config config.Config, writer io.Writer, globalIndex *int) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Warnf("Warning: Skipping file %s due to error: %v", filePath, err)
		return nil
	}

	if config.ClaudeXML {
		xmlOutput := fmt.Sprintf("<document index=\"%d\">\n<source>%s</source>\n<document_content>\n%s\n</document_content>\n</document>\n",
			*globalIndex, filePath, string(content))
		*globalIndex++
		_, err = writer.Write([]byte(xmlOutput))
	} else {
		output := fmt.Sprintf("%s\n---\n%s\n---\n\n", filePath, string(content))
		_, err = writer.Write([]byte(output))
	}

	return err
}

func Run(config config.Config) error {
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

	if !config.IgnoreGitignore {
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
