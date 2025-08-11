package metrics

import (
	"time"
)

const TICKER_INTERVAL = time.Second * 10
const PEGOUT_MAX_DELAY = time.Minute * 20
const AUTOPEGOUT_WARN_DELAY = time.Hour * 18
const AUTOPEGOUT_CRIT_DELAY = time.Hour * 24
const AUTOPEGOUT_PANIC_DELAY = time.Hour * 48
const TOTAL_VALIDATORS = 100
