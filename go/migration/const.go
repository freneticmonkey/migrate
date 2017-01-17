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
	// Is complete and was forced by Continuous Integration
	ForcedCI
	// Is currently being applied
	InProgress
	// Failed to apply
	Failed
	// Skipped application
	Skipped
	// Has been rolled back
	Rollback
)

var StatusString = [10]string{
	"Unapproved",
	"Denied",
	"Depreciated",
	"Approved",
	"Complete",
	"ForcedCI",
	"InProgress",
	"Failed",
	"Skipped",
	"Rollback",
}
