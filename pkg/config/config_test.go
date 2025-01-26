package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestGetEnvVars(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		envVars        map[string]string
		expectedConfig Config
		setupFunc      func()
		cleanupFunc    func()
	}{
		{
			name: "All Environment Variables Set",
			envVars: map[string]string{
				"PATHS":            "path1,path2",
				"EXTENSIONS":       ".go,.py",
				"INCLUDE_HIDDEN":   "true",
				"IGNORE_GITIGNORE": "false",
				"IGNORE_PATTERNS":  "*.log,temp*",
				"OUTPUT_FILE":      "output.txt",
				"CLAUDE_XML":       "true",
			},
			expectedConfig: Config{
				Paths:           []string{"path1", "path2"},
				Extensions:      []string{".go", ".py"},
				IncludeHidden:   true,
				IgnoreGitignore: false,
				IgnorePatterns:  []string{"*.log", "temp*"},
				OutputFile:      "output.txt",
				ClaudeXML:       true,
			},
			setupFunc: func() {
				viper.Reset()
				os.Clearenv()
			},
			cleanupFunc: func() {
				viper.Reset()
				os.Clearenv()
			},
		},
		{
			name:    "No Environment Variables Set",
			envVars: map[string]string{},
			expectedConfig: Config{
				Paths:           []string{},
				Extensions:      []string{},
				IncludeHidden:   false,
				IgnoreGitignore: false,
				IgnorePatterns:  []string{},
				OutputFile:      "",
				ClaudeXML:       false,
			},
			setupFunc: func() {
				viper.Reset()
				os.Clearenv()
			},
			cleanupFunc: func() {
				viper.Reset()
				os.Clearenv()
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			if tc.setupFunc != nil {
				tc.setupFunc()
			}

			// Set environment variables
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}

			// Configure Viper
			viper.AutomaticEnv()

			// Call function
			result := GetEnvVars()

			// Verify
			if len(result.Paths) != len(tc.expectedConfig.Paths) {
				t.Errorf("Paths mismatch. Got %v, want %v", result.Paths, tc.expectedConfig.Paths)
			}
			if len(result.Extensions) != len(tc.expectedConfig.Extensions) {
				t.Errorf("Extensions mismatch. Got %v, want %v", result.Extensions, tc.expectedConfig.Extensions)
			}
			if result.IncludeHidden != tc.expectedConfig.IncludeHidden {
				t.Errorf("IncludeHidden mismatch. Got %v, want %v", result.IncludeHidden, tc.expectedConfig.IncludeHidden)
			}
			if result.IgnoreGitignore != tc.expectedConfig.IgnoreGitignore {
				t.Errorf("IgnoreGitignore mismatch. Got %v, want %v", result.IgnoreGitignore, tc.expectedConfig.IgnoreGitignore)
			}
			if len(result.IgnorePatterns) != len(tc.expectedConfig.IgnorePatterns) {
				t.Errorf("IgnorePatterns mismatch. Got %v, want %v", result.IgnorePatterns, tc.expectedConfig.IgnorePatterns)
			}
			if result.OutputFile != tc.expectedConfig.OutputFile {
				t.Errorf("OutputFile mismatch. Got %v, want %v", result.OutputFile, tc.expectedConfig.OutputFile)
			}
			if result.ClaudeXML != tc.expectedConfig.ClaudeXML {
				t.Errorf("ClaudeXML mismatch. Got %v, want %v", result.ClaudeXML, tc.expectedConfig.ClaudeXML)
			}

			// Cleanup
			if tc.cleanupFunc != nil {
				tc.cleanupFunc()
			}
		})
	}
}

func TestGetEnvVarsWithDotEnvFile(t *testing.T) {
	// Create a temporary .env file
	envContent := `PATHS=test1,test2
EXTENSIONS=.txt
INCLUDE_HIDDEN=true`

	err := os.WriteFile(".env", []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}
	defer os.Remove(".env")

	// Reset Viper and environment
	viper.Reset()
	os.Clearenv()

	// Call function
	result := GetEnvVars()

	// Verify .env file parsing
	if len(result.Paths) != 2 || result.Paths[0] != "test1" || result.Paths[1] != "test2" {
		t.Errorf("Paths from .env file incorrect. Got %v", result.Paths)
	}
	if len(result.Extensions) != 1 || result.Extensions[0] != ".txt" {
		t.Errorf("Extensions from .env file incorrect. Got %v", result.Extensions)
	}
	if !result.IncludeHidden {
		t.Error("IncludeHidden should be true from .env file")
	}
}
