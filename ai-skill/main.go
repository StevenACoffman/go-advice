// distill walks a directory tree, finds all .md source files, and for each
// produces a filled-in copy of distill_source_prompt.md that tells Claude to
// distill that source into a rules file.
//
// Usage: distill [-template PATH] [-dry-run] <directory> <promptout>
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const templateFilename = "distill_source_prompt.md"

const (
	sourcePlaceholder = "<source>{{SOURCE_CONTENT}}</source>"
	destPlaceholder   = "<destination>{{DESTINATION_CONTENT}}</destination>"
)

// Compiled once at startup; used in every call to rulesFilename/filenameToTitle.
var (
	reWordSep    = regexp.MustCompile(`[\s\-]+`)
	reAllWordSep = regexp.MustCompile(`[\s_\-]+`)
)

func main() {
	templateFlag := flag.String("template", "", "path to the prompt template (default: "+templateFilename+" beside the binary, then cwd)")
	dryRun := flag.Bool("dry-run", false, "print what would be written without writing anything")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: distill [-template PATH] [-dry-run] <directory> <promptout>")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	root := flag.Arg(0)
	promptOutDir := flag.Arg(1)

	templatePath, err := resolveTemplate(*templateFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	absTemplate, err := filepath.Abs(templatePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot resolve template path:", err)
		os.Exit(1)
	}

	tmpl, err := os.ReadFile(absTemplate)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot read template:", err)
		os.Exit(1)
	}
	templateStr := string(tmpl)

	if !strings.Contains(templateStr, sourcePlaceholder) {
		fmt.Fprintf(os.Stderr, "template %s does not contain %s\n", absTemplate, sourcePlaceholder)
		os.Exit(1)
	}
	if !strings.Contains(templateStr, destPlaceholder) {
		fmt.Fprintf(os.Stderr, "template %s does not contain %s\n", absTemplate, destPlaceholder)
		os.Exit(1)
	}

	absPromptOut, err := filepath.Abs(promptOutDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot resolve promptout directory:", err)
		os.Exit(1)
	}
	fmt.Printf("promptout dir: %s\n", absPromptOut)

	if !*dryRun {
		if err := os.MkdirAll(absPromptOut, 0755); err != nil {
			fmt.Fprintln(os.Stderr, "cannot create promptout directory:", err)
			os.Exit(1)
		}
		// Verify the directory is reachable after creation.
		if fi, err := os.Stat(absPromptOut); err != nil {
			fmt.Fprintf(os.Stderr, "promptout directory not reachable after MkdirAll: %v\n", err)
			os.Exit(1)
		} else if !fi.IsDir() {
			fmt.Fprintf(os.Stderr, "promptout path exists but is not a directory: %s\n", absPromptOut)
			os.Exit(1)
		}
	}

	var errs []error
	written := 0

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil // keep walking
		}
		if info.IsDir() {
			// Skip hidden directories (e.g. .git), but never skip the root itself.
			if path != root && strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip rules files and previously generated prompt files.
		if strings.HasSuffix(path, "_rules.md") || strings.HasSuffix(path, "_prompt.md") {
			return nil
		}

		// Skip the template itself.
		absPath, err := filepath.Abs(path)
		if err != nil {
			errs = append(errs, err)
			return nil
		}
		if absPath == absTemplate {
			return nil
		}

		if err := processFile(absPath, templateStr, absPromptOut, *dryRun); err != nil {
			errs = append(errs, err)
			return nil // keep walking
		}
		written++
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "error:", e)
		}
		os.Exit(1)
	}

	action := "wrote"
	if *dryRun {
		action = "would write"
	}
	fmt.Printf("%s %d file(s)\n", action, written)
}

// resolveTemplate returns the path to the prompt template. It checks, in order:
//  1. The value of the -template flag (if set).
//  2. templateFilename beside the resolved binary.
//  3. templateFilename in the current working directory.
func resolveTemplate(flagVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}

	// Try beside the binary, resolving symlinks so that `go install` works.
	if exe, err := os.Executable(); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			candidate := filepath.Join(filepath.Dir(resolved), templateFilename)
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}
	}

	// Fall back to the current working directory (convenient when using go run).
	if cwd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(cwd, templateFilename)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf(
		"template %q not found beside binary or in cwd; use -template to specify its location",
		templateFilename,
	)
}

// processFile receives absolute paths for both absSource and absPromptOut —
// no filepath.Abs calls are needed inside.
func processFile(absSource, templateStr, absPromptOut string, dryRun bool) error {
	title, err := extractTitle(absSource)
	if err != nil {
		return fmt.Errorf("extracting title from %s: %w", absSource, err)
	}

	sourceBase := filepath.Base(absSource)
	sourceDir := filepath.Dir(absSource)

	// The destination rules file stays beside the source.
	// Its link is written into the template so Claude knows where to save output.
	destBase := rulesFilename(sourceBase)
	destPath := filepath.Join(sourceDir, destBase)

	// The prompt output file lands in absPromptOut with a _prompt suffix.
	promptPath := filepath.Join(absPromptOut, promptFilename(sourceBase))

	// Links are relative to the prompt file's location (absPromptOut).
	sourceRel, err := filepath.Rel(absPromptOut, absSource)
	if err != nil {
		return fmt.Errorf("computing relative path to source: %w", err)
	}
	destRel, err := filepath.Rel(absPromptOut, destPath)
	if err != nil {
		return fmt.Errorf("computing relative path to dest: %w", err)
	}

	sourceLink := fmt.Sprintf("[%s](%s)", title, filepath.ToSlash(sourceRel))
	destLink := fmt.Sprintf("[%s Rules](%s)", title, filepath.ToSlash(destRel))

	result := strings.ReplaceAll(templateStr, sourcePlaceholder, sourceLink)
	result = strings.ReplaceAll(result, destPlaceholder, destLink)

	if dryRun {
		fmt.Printf("would write %s  (source: %q, dest: %q)\n", promptPath, sourceLink, destLink)
		return nil
	}

	if err := os.WriteFile(promptPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", promptPath, err)
	}

	fi, err := os.Stat(promptPath)
	if err != nil {
		return fmt.Errorf("stat after write (file may not have landed): %w", err)
	}
	fmt.Printf("wrote %s (%d bytes)\n", promptPath, fi.Size())
	return nil
}

// extractTitle returns the text of the first H1 heading in the file, or a
// title-cased form of the filename stem if no heading is found.
func extractTitle(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256*1024), 256*1024) // handle unusually long lines
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# "), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanning %s: %w", path, err)
	}

	// Fallback: derive a readable title from the filename stem.
	stem := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return filenameToTitle(stem), nil
}

// promptFilename derives the prompt output filename from a source filename.
//
//	"self-refinement.md" → "self-refinement_prompt.md"
func promptFilename(name string) string {
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext) + "_prompt" + ext
}

// rulesFilename derives the destination rules filename from a source filename.
//
//	"My-Source File.md" → "my_source_file_rules.md"
func rulesFilename(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	base = strings.ToLower(base)
	base = reWordSep.ReplaceAllString(base, "_")
	return base + "_rules" + ext
}

// filenameToTitle converts a filename stem into a readable title.
//
//	"my-source_file" → "My Source File"
func filenameToTitle(stem string) string {
	words := reAllWordSep.Split(stem, -1)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
