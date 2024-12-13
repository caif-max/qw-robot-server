package common

const (
	MINUTE    = 60
	HOUR      = MINUTE * 60
	HALF_HOUR = MINUTE * 30
	DAY       = HOUR * 24

	// REDIS 前缀
	APP_SESSION_PREFIX = "appSession."
	SessionPrefix      = "QWRobotSession."
)
