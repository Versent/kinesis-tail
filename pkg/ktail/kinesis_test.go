package ktail

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimestamp(t *testing.T) {
	timestamp := int64(1526626158315)

	ts := buildTimestamp(timestamp)

	require.Equal(t, int64(1526626158315000000), ts.UnixNano())
	require.Equal(t, "2018-05-18T06:49:18.315Z", ts.UTC().Format(time.RFC3339Nano))
}
