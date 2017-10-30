package logdata

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/versent/kinesis-tail/pkg/ktail"
	"github.com/versent/kinesis-tail/pkg/matcher"
)

// UncompressLogs uncompress and parse cloudwatch log batch data
func UncompressLogs(includes []string, excludes []string, ts *time.Time, data []byte) ([]*ktail.LogMessage, error) {

	dataReader := bytes.NewReader(data)

	gzipReader, err := gzip.NewReader(dataReader)
	if err != nil {
		return nil, errors.Wrap(err, "un gzip data failed")
	}

	// io.Copy(os.Stdout, gzipReader)
	var batch ktail.LogBatch

	err = json.NewDecoder(gzipReader).Decode(&batch)
	if err != nil {
		return nil, errors.Wrap(err, "json decode failed")
	}

	if !matcher.MatchesTokens(includes, batch.LogGroup, true) {
		return []*ktail.LogMessage{}, nil
	}

	if matcher.MatchesTokens(excludes, batch.LogGroup, false) {
		return []*ktail.LogMessage{}, nil
	}

	logEvents := make([]*ktail.LogMessage, len(batch.LogEvents))

	for i, entry := range batch.LogEvents {
		logEvents[i] = &ktail.LogMessage{
			LogGroup:  batch.LogGroup,
			Timestamp: ts.Format(time.RFC3339),
			Message:   strings.TrimSuffix(entry.Message, "\n"),
		}
	}

	return logEvents, nil
}
