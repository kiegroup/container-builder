package builder

import (
	"context"
	"time"

	"github.com/jpillora/backoff"
	"github.com/ricardozanini/container-builder/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newErrorRecoveryAction() Action {
	// TODO: externalize options
	return &errorRecoveryAction{
		backOff: backoff.Backoff{
			Min:    5 * time.Second,
			Max:    1 * time.Minute,
			Factor: 2,
			Jitter: false,
		},
	}
}

type errorRecoveryAction struct {
	baseAction
	backOff backoff.Backoff
}

func (action *errorRecoveryAction) Name() string {
	return "error-recovery"
}

func (action *errorRecoveryAction) CanHandle(build *api.Build) bool {
	return build.Status.Phase == api.BuildPhaseFailed
}

func (action *errorRecoveryAction) Handle(ctx context.Context, build *api.Build) (*api.Build, error) {
	if build.Status.Failure == nil {
		build.Status.Failure = &api.Failure{
			Reason: build.Status.Error,
			Time:   metav1.Now(),
			Recovery: api.FailureRecovery{
				Attempt:    0,
				AttemptMax: 5,
			},
		}
		return build, nil
	}

	if build.Status.Failure.Recovery.Attempt >= build.Status.Failure.Recovery.AttemptMax {
		build.Status.Phase = api.BuildPhaseError
		return build, nil
	}

	lastAttempt := build.Status.Failure.Recovery.AttemptTime.Time
	if lastAttempt.IsZero() {
		lastAttempt = build.Status.Failure.Time.Time
	}

	elapsed := time.Since(lastAttempt).Seconds()
	elapsedMin := action.backOff.ForAttempt(float64(build.Status.Failure.Recovery.Attempt)).Seconds()

	if elapsed < elapsedMin {
		return nil, nil
	}

	build.Status.Phase = api.BuildPhaseInitialization
	build.Status.Failure.Recovery.Attempt++
	build.Status.Failure.Recovery.AttemptTime = metav1.Now()

	action.L.Infof("Recovery attempt (%d/%d)",
		build.Status.Failure.Recovery.Attempt,
		build.Status.Failure.Recovery.AttemptMax,
	)

	return build, nil
}
