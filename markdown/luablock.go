package markdown

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/speedata/go-lua"
)

// luaBlockPattern matches ```{lua} ... ``` fenced code blocks.
var luaBlockPattern = regexp.MustCompile("(?s)```\\{lua\\}\n(.*?)```")

// inlineExprPattern matches {= expr =} inline expressions.
var inlineExprPattern = regexp.MustCompile(`\{=\s*(.*?)\s*=\}`)

// extractAndRunLuaBlocks finds all ```{lua} ... ``` blocks, executes them
// in the given Lua state, and removes them from the source. If a Lua block
// returns a string, it is inserted into the output. Lua blocks can use
// startrecording() and stoprecording() to capture and process the text
// between two blocks.
func extractAndRunLuaBlocks(l *lua.State, source string) (string, error) {
	matches := luaBlockPattern.FindAllStringSubmatchIndex(source, -1)
	if len(matches) == 0 {
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

	return result.String(), nil
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
		// Get the result from the top of the stack
		val := lua.CheckString(l, -1)
		l.Pop(1)
		return val
	})
	if luaErr != nil {
		return "", luaErr
	}
	return result, nil
}
