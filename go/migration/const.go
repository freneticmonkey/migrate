package migration

// The status states of the Migration or Migration Step
const (
	// Is waiting to be approved
	Unapproved = iota
	// Has been denied
	Denied
	// Has been depreciated
	Depreciated
	// Has been approved
	Approved
	// Is complete
	Complete
	// Is complete and was forced
	Forced
	// Is currently being applied
	InProgress
	// Failed to apply
	Failed
	// Skipped application
	Skipped
)
