package app

// Bindings exposes the Go structs that should be accessible from the Wails runtime.
func Bindings(app *App) []interface{} {
	return []interface{}{app}
}
