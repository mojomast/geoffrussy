package security

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PathSanitizer validates and sanitizes file paths to prevent directory traversal attacks.
// It ensures all file operations stay within the designated project root directory.
type PathSanitizer struct {
	projectRoot string
}

// NewPathSanitizer creates a new PathSanitizer with the specified project root.
// The project root is converted to an absolute path and cleaned.
// Returns an error if the project root cannot be resolved to an absolute path.
func NewPathSanitizer(projectRoot string) (*PathSanitizer, error) {
	if projectRoot == "" {
		return nil, fmt.Errorf("project root cannot be empty")
	}

	// Convert to absolute path and clean it
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project root to absolute path: %w", err)
	}

	// Clean the path to normalize it (removes .., ., redundant separators)
	absRoot = filepath.Clean(absRoot)

	return &PathSanitizer{
		projectRoot: absRoot,
	}, nil
}

// ValidatePath validates that the given path is safe and within the project root.
// It returns the absolute, cleaned path if validation succeeds, or an error if:
// - The path cannot be resolved to an absolute path
// - The resolved path is outside the project root
// - The path contains directory traversal sequences that escape the project root
//
// The returned path is always an absolute, cleaned path suitable for file operations.
func (ps *PathSanitizer) ValidatePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Convert to absolute path
	var absPath string
	var err error

	if filepath.IsAbs(path) {
		// Already absolute, just clean it
		absPath = filepath.Clean(path)
		
		// If it's an absolute path, it must be checked immediately
		// to ensure it's within the project root
		if !ps.isWithinRoot(absPath) {
			return "", fmt.Errorf("path '%s' is outside project root '%s'", path, ps.projectRoot)
		}
	} else {
		// Relative path - resolve relative to project root
		absPath, err = filepath.Abs(filepath.Join(ps.projectRoot, path))
		if err != nil {
			return "", fmt.Errorf("failed to resolve path to absolute: %w", err)
		}
		absPath = filepath.Clean(absPath)
		
		// Check if the resolved path is within the project root
		if !ps.isWithinRoot(absPath) {
			return "", fmt.Errorf("path '%s' is outside project root '%s'", path, ps.projectRoot)
		}
	}

	// Additional check: ensure the path doesn't contain ".." after resolution
	// This catches edge cases where symbolic links or other mechanisms might be used
	if strings.Contains(absPath, "..") {
		return "", fmt.Errorf("path '%s' contains directory traversal sequences", path)
	}

	return absPath, nil
}

// IsPathSafe performs a boolean check to determine if a path is safe.
// It returns true if the path passes validation, false otherwise.
// This is a convenience method that wraps ValidatePath for cases where
// only a boolean result is needed without error details.
func (ps *PathSanitizer) IsPathSafe(path string) bool {
	_, err := ps.ValidatePath(path)
	return err == nil
}

// isWithinRoot checks if the given absolute path is within the project root.
// It uses filepath.Rel to determine the relationship between the paths.
// A path is considered within the root if:
// - The relative path doesn't start with ".."
// - The relative path is not "." (which would mean they're the same)
func (ps *PathSanitizer) isWithinRoot(absPath string) bool {
	// Use filepath.Rel to get the relative path from project root to the target
	rel, err := filepath.Rel(ps.projectRoot, absPath)
	if err != nil {
		return false
	}

	// If the relative path starts with "..", it's outside the root
	if strings.HasPrefix(rel, "..") {
		return false
	}

	// If the relative path starts with ".." after converting to forward slashes
	// (handles Windows paths)
	relForward := filepath.ToSlash(rel)
	if strings.HasPrefix(relForward, "../") || relForward == ".." {
		return false
	}

	return true
}

// GetProjectRoot returns the project root path used by this sanitizer.
func (ps *PathSanitizer) GetProjectRoot() string {
	return ps.projectRoot
}
