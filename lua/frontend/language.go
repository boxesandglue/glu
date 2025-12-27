package frontend

import (
	"github.com/boxesandglue/boxesandglue/backend/lang"
	"github.com/speedata/go-lua"
)

const languageMetaTable = "Language"

// Language wraps the boxesandglue lang.Lang type
type Language struct {
	Value *lang.Lang
}

// checkLanguage retrieves a Language userdata from the stack
func checkLanguage(l *lua.State, index int) *Language {
	ud := lua.CheckUserData(l, index, languageMetaTable)
	if lang, ok := ud.(*Language); ok {
		return lang
	}
	lua.Errorf(l, "Language expected")
	return nil
}

// languageIndex handles attribute access (__index metamethod)
func languageIndex(l *lua.State) int {
	lang := checkLanguage(l, 1)
	key := lua.CheckString(l, 2)

	switch key {
	case "name":
		l.PushString(lang.Value.Name)
		return 1
	}

	return 0
}

// registerLanguageMetaTable creates the Language metatable
func registerLanguageMetaTable(l *lua.State) {
	lua.NewMetaTable(l, languageMetaTable)
	lua.SetFunctions(l, []lua.RegistryFunction{
		{Name: "__index", Function: languageIndex},
	}, 0)
	l.Pop(1)
}
