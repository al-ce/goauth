package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/al-ce/goauth/internal/repository"
	"github.com/al-ce/goauth/pkg/config"
)

// StartJobs starts jobs with a context from main
func StartJobs(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB) {

	// NOTE: Add goroutines here for any future jobs

	wg.Add(1)
	go func() {
		defer wg.Done()
		UnlockExpiredLocks(ctx, config.AccountUnlockPeriod, db)
	}()
}

// UnlockExpiredLocks calls the repo method to unlock all accounts whose
// locks have expired every `period`
func UnlockExpiredLocks(
	ctx context.Context,
	period time.Duration,
	db *gorm.DB,
) {
	ur, err := repository.NewUserRepository(db)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("[Jobs] [ERROR] Could not init user repo: %s", err.Error()))
		return
	}
	ticker := time.NewTicker(period)
	defer ticker.Stop()

	// Perform an initial check before starting the ticker
	unlockHelper(ur)

	for {
		select {
		case <-ticker.C:
			unlockHelper(ur)
		case <-ctx.Done():
			log.Info().Msg("[Jobs] [UnlockExpiredLocks] Stopping job")
			return
		}
	}
}

func unlockHelper(ur *repository.UserRepository) {
	affected, err := ur.UnlockAllExpiredLocks()
	if err != nil {
		log.Error().Msg(fmt.Sprintf("[Jobs] [ERROR] %s", err.Error()))
	} else {
		log.Info().Msg(fmt.Sprintf("[Jobs] [UnlockExpiredLocks] %d rows affected", affected))
	}
}
