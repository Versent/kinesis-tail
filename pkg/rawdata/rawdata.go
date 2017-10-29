package rawdata

import (
	"strings"
	"time"

	"github.com/versent/kinesis-tail/pkg/ktail"
)

// DecodeRawData format the raw data
func DecodeRawData(ts *time.Time, data []byte) *ktail.LogMessage {
	return &ktail.LogMessage{
		Timestamp: ts.Format(time.RFC3339),
		Message:   strings.TrimSuffix(string(data), "\n"),
	}
}
