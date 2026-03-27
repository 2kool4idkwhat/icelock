package main

import (
	"icelock/log"
	"os"

	"kernel.org/pub/linux/libs/security/libcap/cap"
)

func setupCaps(cfg *config) {

	if !cfg.KeepCaps {
		old := cap.GetProc()
		empty := cap.NewSet()

		log.Debug("Current capabilities: %q", old)

		if old.String() == empty.String() {
			log.Info("No capabilities to drop, skipping")
			return
		}

		err := empty.SetProc()
		if err != nil {
			log.Error("Failed to drop capabilities: %v", err)
			os.Exit(1)
		}

		log.Info("Dropped capabilities")
	}

}
