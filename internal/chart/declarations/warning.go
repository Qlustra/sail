package declarations

// Warning describes a non-fatal grammar compilation advisory.
type Warning struct {
	// Code identifies the advisory in a machine-readable form.
	Code string
	// Message explains the advisory to a reader.
	Message string
}
