package configsetup

import "github.com/freneticmonkey/migrate/go/util"

// HealthCheck Stores basic information about a check
type HealthCheck struct {
	Details string
	Result  int
}

// Health Stores the result of all checks
type Health struct {
	Checks  []HealthCheck
	success bool
}

// GetHealth Get a Health instance
func GetHealth() Health {
	return Health{
		success: true,
	}
}

// AddCheck Add a HealthCheck
func (h *Health) AddCheck(hc HealthCheck) {
	h.Checks = append(h.Checks, hc)
	if hc.Result != 0 {
		h.success = false
	}
}

// AddPass Helper function for adding a passing check
func (h *Health) AddPass(details string) {
	h.AddCheck(HealthCheck{
		Details: details,
		Result:  0,
	})
}

// AddFail Helper function for adding a failing check
func (h *Health) AddFail(details string) {
	h.AddCheck(HealthCheck{
		Details: details,
		Result:  1,
	})
}

// AddFailValue Helper function for adding a failing check with a fail value
func (h *Health) AddFailValue(details string, result int) {
	h.AddCheck(HealthCheck{
		Details: details,
		Result:  result,
	})
}

// Ok All checks passing
func (h Health) Ok() bool {
	return h.success
}

// Display Print health checks to the console
func (h Health) Display() {
	for _, check := range h.Checks {
		if check.Result != 0 {
			util.LogError(check.Details)
		} else {
			util.LogOk(check.Details)
		}
	}
}
