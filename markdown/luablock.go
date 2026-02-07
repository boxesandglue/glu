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
// in the given Lua state, and removes them from the source.
func extractAndRunLuaBlocks(l *lua.State, source string) (string, error) {
	var luaErr error
	result := luaBlockPattern.ReplaceAllStringFunc(source, func(match string) string {
		if luaErr != nil {
			return ""
		}
		// Extract the Lua code between ```{lua}\n and ```
		sub := luaBlockPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return ""
		}
		code := strings.TrimSpace(sub[1])
		if err := lua.DoString(l, code); err != nil {
			luaErr = fmt.Errorf("lua block error: %w", err)
		}
		return ""
	})
	if luaErr != nil {
		return "", luaErr
	}
	return result, nil
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
