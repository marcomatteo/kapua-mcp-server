package handlers

import "time"

// timeNow is wrapped for tests so we can stub the clock where needed.
var timeNow = time.Now
