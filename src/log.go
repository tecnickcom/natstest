package main

import (
	"log"
	"os"
)

// syslog error levels
var (
	// 0 - Emergency: System is unusable
	emergLog = log.New(os.Stderr, "[EMERG] [NATSTEST] ", log.LstdFlags|log.Lshortfile|log.LUTC)

	// 1 - Alert: Should be corrected immediately
	alertLog = log.New(os.Stderr, "[ALERT] [NATSTEST] ", log.LstdFlags|log.Lshortfile|log.LUTC)

	// 2 - Critical: Critical conditions
	critLog = log.New(os.Stderr, "[CRIT] [NATSTEST] ", log.LstdFlags|log.Lshortfile|log.LUTC)

	// 3 - Error: Error conditions
	errLog = log.New(os.Stderr, "[ERR] [NATSTEST] ", log.LstdFlags|log.Lshortfile|log.LUTC)

	// 4 - Warning: May indicate that an error will occur if action is not taken.
	warningLog = log.New(os.Stderr, "[WARNING] [NATSTEST] ", log.LstdFlags|log.LUTC)

	// 5 - Notice: Events that are unusual, but not error conditions.
	noticeLog = log.New(os.Stderr, "[NOTICE] [NATSTEST] ", log.LstdFlags|log.LUTC)

	// 6 - Informational: Normal operational messages that require no action.
	infoLog = log.New(os.Stdout, "[INFO] [NATSTEST] ", log.LstdFlags|log.LUTC)

	// 7 - Debug: Information useful to developers for debugging the application.
	debugLog = log.New(os.Stderr, "[DEBUG] [NATSTEST] ", log.LstdFlags|log.Lshortfile|log.LUTC)
)
