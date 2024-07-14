package contexts

type ContextKey string

const (
	// DiagnosticLogContextKey is the key for the diagnostic log context.
	DiagnosticLogContextKey ContextKey = "diagnostics"
	// ActivityLogContextKey is the key for the activity log context.
	ActivityLogContextKey ContextKey = "activity"
)
