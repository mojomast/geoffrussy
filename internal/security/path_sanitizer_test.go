package security

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// getAbsolutePathOutsideRoot returns an absolute path that's outside the test project root
func getAbsolutePathOutsideRoot() string {
	if runtime.GOOS == "windows" {
		return "C:\\Windows\\System32\\config"
	}
	return "/etc/passwd"
}

func TestNewPathSanitizer(t *testing.T) {
	tests := []struct {
		name        string
		projectRoot string
		wantErr     bool
		errContains string
	}{
		{
			name:        "valid absolute path",
			projectRoot: "/home/user/project",
			wantErr:     false,
		},
		{
			name:        "valid relative path",
			projectRoot: ".",
			wantErr:     false,
		},
		{
			name:        "empty path",
			projectRoot: "",
			wantErr:     true,
			errContains: "project root cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps, err := NewPathSanitizer(tt.projectRoot)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewPathSanitizer() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("NewPathSanitizer() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("NewPathSanitizer() unexpected error = %v", err)
				return
			}
			if ps == nil {
				t.Errorf("NewPathSanitizer() returned nil")
				return
			}
			// Verify project root is absolute
			if !filepath.IsAbs(ps.projectRoot) {
				t.Errorf("NewPathSanitizer() project root is not absolute: %v", ps.projectRoot)
			}
		})
	}
}

func TestPathSanitizer_ValidatePath(t *testing.T) {
	// Create a temporary project root for testing
	projectRoot := "/home/user/project"
	if runtime.GOOS == "windows" {
		projectRoot = "C:\\Users\\user\\project"
	}

	ps, err := NewPathSanitizer(projectRoot)
	if err != nil {
		t.Fatalf("Failed to create PathSanitizer: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid relative path",
			path:    "src/main.go",
			wantErr: false,
		},
		{
			name:    "valid relative path with subdirectory",
			path:    "internal/security/path_sanitizer.go",
			wantErr: false,
		},
		{
			name:    "current directory",
			path:    ".",
			wantErr: false,
		},
		{
			name:        "directory traversal with ../",
			path:        "../../../etc/passwd",
			wantErr:     true,
			errContains: "outside project root",
		},
		{
			name:        "directory traversal with ..",
			path:        "..",
			wantErr:     true,
			errContains: "outside project root",
		},
		{
			name:        "absolute path outside root",
			path:        getAbsolutePathOutsideRoot(),
			wantErr:     true,
			errContains: "outside project root",
		},
		{
			name:        "empty path",
			path:        "",
			wantErr:     true,
			errContains: "path cannot be empty",
		},
		{
			name:    "path with redundant separators",
			path:    "src//main.go",
			wantErr: false,
		},
		{
			name:    "path with . in middle",
			path:    "src/./main.go",
			wantErr: false,
		},
		{
			name:    "path with .. that stays within root",
			path:    "src/../main.go",
			wantErr: false,
		},
	}

	// Add Windows-specific tests
	if runtime.GOOS == "windows" {
		tests = append(tests, []struct {
			name        string
			path        string
			wantErr     bool
			errContains string
		}{
			{
				name:        "Windows absolute path outside root",
				path:        "C:\\Windows\\System32\\config",
				wantErr:     true,
				errContains: "outside project root",
			},
			{
				name:        "Windows UNC path",
				path:        "\\\\server\\share\\file.txt",
				wantErr:     true,
				errContains: "outside project root",
			},
			{
				name:        "Windows directory traversal",
				path:        "..\\..\\..\\Windows\\System32",
				wantErr:     true,
				errContains: "outside project root",
			},
		}...)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			safePath, err := ps.ValidatePath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidatePath() expected error but got none, returned path: %v", safePath)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidatePath() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("ValidatePath() unexpected error = %v", err)
				return
			}
			// Verify returned path is absolute
			if !filepath.IsAbs(safePath) {
				t.Errorf("ValidatePath() returned non-absolute path: %v", safePath)
			}
			// Verify returned path is within project root
			if !strings.HasPrefix(safePath, ps.projectRoot) {
				t.Errorf("ValidatePath() returned path %v not within project root %v", safePath, ps.projectRoot)
			}
		})
	}
}

func TestPathSanitizer_IsPathSafe(t *testing.T) {
	projectRoot := "/home/user/project"
	if runtime.GOOS == "windows" {
		projectRoot = "C:\\Users\\user\\project"
	}

	ps, err := NewPathSanitizer(projectRoot)
	if err != nil {
		t.Fatalf("Failed to create PathSanitizer: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantSafe bool
	}{
		{
			name:     "valid relative path",
			path:     "src/main.go",
			wantSafe: true,
		},
		{
			name:     "directory traversal",
			path:     "../../../etc/passwd",
			wantSafe: false,
		},
		{
			name:     "absolute path outside root",
			path:     getAbsolutePathOutsideRoot(),
			wantSafe: false,
		},
		{
			name:     "empty path",
			path:     "",
			wantSafe: false,
		},
		{
			name:     "current directory",
			path:     ".",
			wantSafe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isSafe := ps.IsPathSafe(tt.path)
			if isSafe != tt.wantSafe {
				t.Errorf("IsPathSafe() = %v, want %v", isSafe, tt.wantSafe)
			}
		})
	}
}

func TestPathSanitizer_GetProjectRoot(t *testing.T) {
	projectRoot := "/home/user/project"
	ps, err := NewPathSanitizer(projectRoot)
	if err != nil {
		t.Fatalf("Failed to create PathSanitizer: %v", err)
	}

	root := ps.GetProjectRoot()
	if !filepath.IsAbs(root) {
		t.Errorf("GetProjectRoot() returned non-absolute path: %v", root)
	}
	// The root should be cleaned and absolute
	expectedRoot, _ := filepath.Abs(projectRoot)
	expectedRoot = filepath.Clean(expectedRoot)
	if root != expectedRoot {
		t.Errorf("GetProjectRoot() = %v, want %v", root, expectedRoot)
	}
}

// TestPathSanitizer_EdgeCases tests various edge cases
func TestPathSanitizer_EdgeCases(t *testing.T) {
	projectRoot := "/home/user/project"
	if runtime.GOOS == "windows" {
		projectRoot = "C:\\Users\\user\\project"
	}

	ps, err := NewPathSanitizer(projectRoot)
	if err != nil {
		t.Fatalf("Failed to create PathSanitizer: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		wantErr     bool
		description string
	}{
		{
			name:        "multiple slashes",
			path:        "src///main.go",
			wantErr:     false,
			description: "Multiple slashes should be normalized",
		},
		{
			name:        "trailing slash",
			path:        "src/",
			wantErr:     false,
			description: "Trailing slash should be handled",
		},
		{
			name:        "dot segments in middle",
			path:        "src/./internal/./main.go",
			wantErr:     false,
			description: "Dot segments should be normalized",
		},
		{
			name:        "parent directory within bounds",
			path:        "src/internal/../main.go",
			wantErr:     false,
			description: "Parent directory that stays within root should be allowed",
		},
		{
			name:        "complex traversal attempt",
			path:        "src/../../../../../../etc/passwd",
			wantErr:     true,
			description: "Complex traversal should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ps.ValidatePath(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("ValidatePath() expected error for %s but got none", tt.description)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePath() unexpected error for %s: %v", tt.description, err)
			}
		})
	}
}

// TestPathSanitizer_PathNormalization tests that equivalent paths are normalized consistently
func TestPathSanitizer_PathNormalization(t *testing.T) {
	projectRoot := "/home/user/project"
	if runtime.GOOS == "windows" {
		projectRoot = "C:\\Users\\user\\project"
	}

	ps, err := NewPathSanitizer(projectRoot)
	if err != nil {
		t.Fatalf("Failed to create PathSanitizer: %v", err)
	}

	// Test that different representations of the same path normalize to the same result
	paths := []string{
		"src/main.go",
		"src//main.go",
		"src/./main.go",
		"./src/main.go",
	}

	var normalizedPaths []string
	for _, path := range paths {
		safePath, err := ps.ValidatePath(path)
		if err != nil {
			t.Errorf("ValidatePath(%s) unexpected error: %v", path, err)
			continue
		}
		normalizedPaths = append(normalizedPaths, safePath)
	}

	// All normalized paths should be identical
	if len(normalizedPaths) > 1 {
		firstPath := normalizedPaths[0]
		for i, path := range normalizedPaths[1:] {
			if path != firstPath {
				t.Errorf("Path normalization inconsistent: paths[0]=%s, paths[%d]=%s", firstPath, i+1, path)
			}
		}
	}
}

// TestPathSanitizer_SpecialCharacters tests paths with special characters
func TestPathSanitizer_SpecialCharacters(t *testing.T) {
	projectRoot := "/home/user/project"
	if runtime.GOOS == "windows" {
		projectRoot = "C:\\Users\\user\\project"
	}

	ps, err := NewPathSanitizer(projectRoot)
	if err != nil {
		t.Fatalf("Failed to create PathSanitizer: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "path with spaces",
			path:    "my folder/my file.txt",
			wantErr: false,
		},
		{
			name:    "path with hyphens",
			path:    "my-folder/my-file.txt",
			wantErr: false,
		},
		{
			name:    "path with underscores",
			path:    "my_folder/my_file.txt",
			wantErr: false,
		},
		{
			name:    "path with dots in filename",
			path:    "src/file.test.go",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ps.ValidatePath(tt.path)
			if tt.wantErr && err == nil {
				t.Errorf("ValidatePath() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ValidatePath() unexpected error: %v", err)
			}
		})
	}
}
