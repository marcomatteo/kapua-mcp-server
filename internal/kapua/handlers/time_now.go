package handlers

import "time"

// timeNow is wrapped for tests so we can inject a deterministic clock.
var timeNow = time.Now
