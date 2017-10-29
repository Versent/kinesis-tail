package sorter

import (
	"io"
	"sort"

	"github.com/versent/kinesis-tail/pkg/ktail"
)

// FormatFunc func which is passed to the message sorter and invoked for each line to format it
type FormatFunc func(wr io.Writer, msg *ktail.LogMessage)

// MessageSorter manages a cache of messages and sorts then formats them on each flush
type MessageSorter struct {
	batchSize int
	current   int
	cache     []*ktail.LogMessage
	wr        io.Writer
	format    FormatFunc
}

// New create a new message sorter
func New(wr io.Writer, batchSize int, formatFunc FormatFunc) *MessageSorter {
	return &MessageSorter{
		batchSize: batchSize,
		wr:        wr,
		format:    formatFunc,
	}
}

// PushBatch this inserts a batch in the cache and checks whether to flush
func (lms *MessageSorter) PushBatch(logMessageBatch []*ktail.LogMessage) bool {
	lms.cache = append(lms.cache, logMessageBatch...)
	return lms.flushCheck()
}

func (lms *MessageSorter) flushCheck() bool {
	lms.current++

	if lms.current != lms.batchSize {
		return false
	}
	sort.Sort(ktail.ByTimestamp(lms.cache))

	for _, msg := range lms.cache {
		lms.format(lms.wr, msg)
	}

	lms.cache = []*ktail.LogMessage{}
	lms.current = 0

	return true
}
