package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalcInformTime(t *testing.T) {
	type args struct {
		periodicInformTime     time.Time
		startedAt              time.Time
		now                    time.Time
		periodicInformEnabled  bool
		periodicInformInterval time.Duration
	}
	calc := func(a args) time.Time {
		return calcInformTime(
			a.periodicInformTime,
			a.startedAt,
			a.now,
			a.periodicInformEnabled,
			a.periodicInformInterval,
		)
	}

	t.Run("Scheduled with PeriodicInformTime", func(t *testing.T) {
		pit := time.Date(2000, 05, 05, 10, 20, 30, 0, time.UTC)
		now := time.Date(2000, 05, 01, 0, 0, 0, 0, time.UTC)
		res := calc(args{
			periodicInformTime: pit,
			now:                now,
		})
		assert.WithinDuration(t, pit, res, 0)
	})
	t.Run("PeriodicInform disabled", func(t *testing.T) {
		pit := time.Date(2000, 05, 05, 10, 20, 30, 0, time.UTC)
		now := pit.Add(1 * time.Hour)
		res := calc(args{
			periodicInformTime:    pit,
			now:                   now,
			periodicInformEnabled: false,
		})
		assert.WithinDuration(t, pit.Add(365*24*time.Hour+1*time.Hour), res, 0)
	})
	t.Run("First inform after PeriodicInformTime", func(t *testing.T) {
		pit := time.Date(2000, 05, 05, 10, 20, 30, 0, time.UTC)
		now := pit.Add(1 * time.Minute)
		res := calc(args{
			periodicInformTime:     pit,
			now:                    now,
			periodicInformEnabled:  true,
			periodicInformInterval: 10 * time.Minute,
		})
		assert.WithinDuration(t, now.Add(9*time.Minute), res, 0)
	})
	t.Run("Third inform after PeriodicInformTime", func(t *testing.T) {
		pit := time.Date(2000, 05, 05, 10, 20, 30, 0, time.UTC)
		now := pit.Add(25 * time.Minute)
		res := calc(args{
			periodicInformTime:     pit,
			now:                    now,
			periodicInformEnabled:  true,
			periodicInformInterval: 10 * time.Minute,
		})
		assert.WithinDuration(t, now.Add(5*time.Minute), res, 0)
	})
	t.Run("Second inform, PeriodicInformTime not set", func(t *testing.T) {
		startedAt := time.Date(2000, 05, 05, 10, 20, 30, 0, time.UTC)
		now := startedAt.Add(5 * time.Minute)
		res := calc(args{
			startedAt:              startedAt,
			now:                    now,
			periodicInformEnabled:  true,
			periodicInformInterval: 10 * time.Minute,
		})
		assert.WithinDuration(t, now.Add(5*time.Minute), res, 0)
	})
}
