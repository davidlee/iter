#!/usr/bin/env bash

# Script to rename Vice to Vice preserving case throughout the codebase

set -e

echo "Renaming Vice to Vice throughout the codebase..."

# Find all text files (excluding binary files, git directories, and common build artifacts)
find . -type f \
  -not -path "./.git/*" \
  -not -path "./vendor/*" \
  -not -path "./node_modules/*" \
  -not -path "./.next/*" \
  -not -path "./dist/*" \
  -not -path "./build/*" \
  -not -name "*.png" \
  -not -name "*.jpg" \
  -not -name "*.jpeg" \
  -not -name "*.gif" \
  -not -name "*.ico" \
  -not -name "*.svg" \
  -not -name "*.webp" \
  -not -name "*.pdf" \
  -not -name "*.zip" \
  -not -name "*.tar.gz" \
  -not -name "*.exe" \
  -not -name "*.dll" \
  -not -name "*.so" \
  -not -name "*.dylib" \
  -not -name "*.a" \
  -not -name "*.o" |
  while read -r file; do
    # Check if file contains text (not binary)
    if file "$file" | grep -q text; then
      # Perform case-preserving replacements
      sed -i.bak \
        -e 's/\bIter\b/Vice/g' \
        -e 's/\biter\b/vice/g' \
        -e 's/\bITER\b/VICE/g' \
        "$file"

      # Remove backup file if no changes were made
      if cmp -s "$file" "$file.bak"; then
        rm "$file.bak"
      else
        echo "Updated: $file"
        rm "$file.bak"
      fi
    fi
  done

echo "Renaming complete. Remember to:"
echo "1. Update any module names in go.mod"
echo "2. Rename directory if needed"
echo "3. Update any CI/CD configurations"
echo "4. Update remote repository name if applicable"
