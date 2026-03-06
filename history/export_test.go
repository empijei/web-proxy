package history

import "time"

func (r *Recorder) SetClock(now func() time.Time) {
	r.now = now
}
