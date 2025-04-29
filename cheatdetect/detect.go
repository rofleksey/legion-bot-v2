package cheatdetect

import (
	"context"
	"fmt"
	"legion-bot-v2/cheatdetect/common"
	"legion-bot-v2/cheatdetect/discord"
	"sync"

	"github.com/hashicorp/go-multierror"
)

type Detector struct {
	detectors []common.Detector
}

func NewDetector() *Detector {
	return &Detector{
		detectors: []common.Detector{
			//unknown.New(),
			discord.New(),
		},
	}
}

func (d *Detector) Detect(ctx context.Context, username string) ([]common.DetectedUser, *multierror.Error) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []common.DetectedUser
		errs    *multierror.Error
	)

	for _, detector := range d.detectors {
		wg.Add(1)

		go func(det common.Detector) {
			defer wg.Done()

			detectedUsers, err := det.Detect(ctx, username)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("%s detector failed: %w", detector.Name(), err))
			}
			if len(detectedUsers) > 0 {
				results = append(results, detectedUsers...)
			}
		}(detector)
	}

	wg.Wait()

	if results == nil {
		results = []common.DetectedUser{}
	}

	return results, errs
}
