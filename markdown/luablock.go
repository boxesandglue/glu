package markdown

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/speedata/go-lua"
)

// luaBlockPattern matches ```{lua} ... ``` fenced code blocks.
var luaBlockPattern = regexp.MustCompile("(?s)```\\{lua\\}\n(.*?)```")

// protectExtendedFences replaces fenced code blocks opened with 4+ backticks
// with placeholders. Go's regexp (RE2) has no backreferences, so we scan manually.
func protectExtendedFences(source string, protected *[]string) string {
	var out strings.Builder
	rest := source
	for {
		// Find a line starting with 4+ backticks.
		idx := strings.Index(rest, "````")
		if idx < 0 {
			break
		}
		// Must be at start of line (or start of string).
		if idx > 0 && rest[idx-1] != '\n' {
			out.WriteString(rest[:idx+4])
			rest = rest[idx+4:]
			continue
		}
		// Count the backticks.
		fence := idx
		end := idx
		for end < len(rest) && rest[end] == '`' {
			end++
		}
		fenceStr := rest[idx:end]
		// Find closing fence (same number of backticks, on its own line).
		closePattern := "\n" + fenceStr
		closeIdx := strings.Index(rest[end:], closePattern)
		if closeIdx < 0 {
			out.WriteString(rest[:end])
			rest = rest[end:]
			continue
		}
		// Include everything up to and including the closing fence.
		blockEnd := end + closeIdx + len(closePattern)
		block := rest[fence:blockEnd]
		placeholder := fmt.Sprintf("\x00PROTECTED_%d\x00", len(*protected))
		*protected = append(*protected, block)
		out.WriteString(rest[:fence])
		out.WriteString(placeholder)
		rest = rest[blockEnd:]
	}
	out.WriteString(rest)
	return out.String()
}

// inlineExprPattern matches {= expr =} inline expressions.
var inlineExprPattern = regexp.MustCompile(`\{=\s*(.*?)\s*=\}`)

// extractAndRunLuaBlocks finds all ```{lua} ... ``` blocks, executes them
// in the given Lua state, and removes them from the source. If a Lua block
// returns a string, it is inserted into the output. Lua blocks can use
// startrecording() and stoprecording() to capture and process the text
// between two blocks.
//
// Blocks marked with ```{!lua} are not executed — the {!lua} tag is
// converted to plain "lua" so goldmark renders them with syntax highlighting.
func extractAndRunLuaBlocks(l *lua.State, source string) (string, error) {
	// Protect fenced code blocks with 4+ backticks from Lua extraction.
	// These may contain ```{lua} examples that should not be executed.
	var protected []string
	source = protectExtendedFences(source, &protected)

	// Convert display-only blocks before matching executable ones.
	source = strings.ReplaceAll(source, "```{!lua}", "```lua")
	matches := luaBlockPattern.FindAllStringSubmatchIndex(source, -1)
	if len(matches) == 0 {
		for i, orig := range protected {
			source = strings.Replace(source, fmt.Sprintf("\x00PROTECTED_%d\x00", i), orig, 1)
		}
		return source, nil
	}

	var result strings.Builder
	var recordBuf strings.Builder
	recording := false

	// Register startrecording() — begins capturing text between Lua blocks.
	l.PushGoFunction(func(l *lua.State) int {
		recording = true
		recordBuf.Reset()
		return 0
	})
	l.SetGlobal("startrecording")

	// Register stoprecording() — stops capturing and returns the captured text.
	l.PushGoFunction(func(l *lua.State) int {
		recording = false
		l.PushString(recordBuf.String())
		recordBuf.Reset()
		return 1
	})
	l.SetGlobal("stoprecording")

	pos := 0
	for _, match := range matches {
		// Text before this Lua block
		between := source[pos:match[0]]
		if recording {
			recordBuf.WriteString(between)
		} else {
			result.WriteString(between)
		}

		// Extract and execute the Lua code
		code := strings.TrimSpace(source[match[2]:match[3]])
		top := l.Top()
		if err := lua.DoString(l, code); err != nil {
			return "", fmt.Errorf("lua block error: %w", err)
		}

		// If the Lua block returned a string, insert it into the output.
		if l.Top() > top && l.IsString(-1) {
			val, _ := l.ToString(-1)
			l.SetTop(top)
			result.WriteString(val)
		}

		pos = match[1]
	}

	// Remaining text after the last Lua block
	remaining := source[pos:]
	if recording {
		recordBuf.WriteString(remaining)
	} else {
		result.WriteString(remaining)
	}

	out := result.String()
	for i, orig := range protected {
		out = strings.Replace(out, fmt.Sprintf("\x00PROTECTED_%d\x00", i), orig, 1)
	}
	return out, nil
}

// expandInlineExpressions replaces {= expr =} with the Lua evaluation result.
func expandInlineExpressions(l *lua.State, source string) (string, error) {
	var luaErr error
	result := inlineExprPattern.ReplaceAllStringFunc(source, func(match string) string {
		if luaErr != nil {
			return match
		}
		sub := inlineExprPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		expr := sub[1]
		// Wrap in "return" so the expression yields a value
		code := "return " + expr
		if err := lua.DoString(l, code); err != nil {
			luaErr = fmt.Errorf("inline expression {= %s =}: %w", expr, err)
			return match
		}
		// Coerce the top-of-stack to a string, matching Lua's
		// own tostring() semantics (numbers / booleans / nil all
		// stringify naturally; tables / functions become a type
		// tag). CheckString here would have errored on numbers,
		// which made `{= 2+2 =}` blow up.
		val, ok := l.ToString(-1)
		if !ok {
			val = fmt.Sprintf("%v", l.ToValue(-1))
		}
		l.Pop(1)
		return val
	})
	if luaErr != nil {
		return "", luaErr
	}
	return result, nil
}
