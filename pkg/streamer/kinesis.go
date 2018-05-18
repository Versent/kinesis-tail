package streamer

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/pkg/errors"
)

// KinesisStreamer this manages streaming data from a number of shards asynchronously
type KinesisStreamer struct {
	svc           kinesisiface.KinesisAPI
	iterators     map[string]*string
	iteratorMutex *sync.Mutex
	pollFreqMs    int64
	logger        *logrus.Logger
}

// GetRecordsEntry returns the results of the last get records request
type GetRecordsEntry struct {
	Created time.Time
	Shard   string
	Records []*kinesis.Record
	Err     error
}

// New return a new configured streamer
func New(svc kinesisiface.KinesisAPI, iterators map[string]*string, pollFreqMs int64, logger *logrus.Logger) *KinesisStreamer {
	return &KinesisStreamer{
		svc:           svc,
		iterators:     iterators,
		pollFreqMs:    pollFreqMs,
		iteratorMutex: &sync.Mutex{},
		logger:        logger,
	}
}

// StartGetRecords intiate the streaming of records using the configured iterators
func (ks *KinesisStreamer) StartGetRecords() chan *GetRecordsEntry {

	ch := make(chan *GetRecordsEntry)

	for key := range ks.iterators {
		go ks.asyncGetRecords(key, ch)
	}

	return ch
}

func (ks *KinesisStreamer) asyncGetRecords(shard string, ch chan *GetRecordsEntry) {

	c := time.Tick(time.Duration(ks.pollFreqMs) * time.Millisecond)

	for now := range c {

		if ks.iterators[shard] == nil {
			ks.logger.Debug("nil iterator for shard as it is CLOSED: %s", shard)
			continue
		}

		resp, err := ks.svc.GetRecords(&kinesis.GetRecordsInput{
			ShardIterator: ks.iterators[shard],
		})
		if err != nil {
			ch <- &GetRecordsEntry{Created: now, Shard: shard, Err: errors.Wrap(err, "get records failed")}
		}

		ks.logger.WithField("iterator", resp).Debug("get records shard")

		ch <- &GetRecordsEntry{Created: now, Shard: shard, Records: resp.Records}

		ks.iteratorMutex.Lock()
		ks.iterators[shard] = resp.NextShardIterator
		ks.iteratorMutex.Unlock()
	}
}
