package history

import "github.com/oklog/ulid/v2"

type Entry struct{}

type Recorder struct {
	state map[ulid.ULID]Entry
}
