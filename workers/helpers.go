package workers

import (
	"math"
	"time"

	"github.com/hashicorp/go-tfe"
)

func isDiscarded(status tfe.RunStatus) bool {
	return status == tfe.RunDiscarded
}

func isCanceled(status tfe.RunStatus) bool {
	return status == tfe.RunCanceled
}

func isErrored(status tfe.RunStatus) bool {
	return status == tfe.RunErrored
}

func isPlanned(status tfe.RunStatus) bool {
	return status == tfe.RunPlanned
}

func isCostEstimated(status tfe.RunStatus) bool {
	return status == tfe.RunCostEstimated
}

func isApplied(status tfe.RunStatus) bool {
	return status == tfe.RunApplied
}

func isConfirmed(status tfe.RunStatus) bool {
	return status == tfe.RunConfirmed
}

func backoff(min, max float64, iter int) time.Duration {
	backoff := math.Pow(2, float64(iter)/5) * min
	if backoff > max {
		backoff = max
	}

	return time.Duration(backoff) * time.Millisecond
}
