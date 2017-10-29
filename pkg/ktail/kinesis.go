package ktail

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// KinesisHelper simple helper for general high level kinesis operations
type KinesisHelper struct {
	svc    kinesisiface.KinesisAPI
	logger *logrus.Logger
}

type iteratorResult struct {
	shardID  string
	iterator *string
}

// New build a new configured kinesis helper
func New(svc kinesisiface.KinesisAPI, logger *logrus.Logger) *KinesisHelper {
	return &KinesisHelper{
		svc:    svc,
		logger: logger,
	}
}

// GetStreamIterators build a list of iterators for the stream
func (kh *KinesisHelper) GetStreamIterators(streamName string) (map[string]*string, error) {

	respDesc, err := kh.svc.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		return nil, errors.Wrap(err, "describe stream failed")
	}
	kh.logger.WithField("respDesc", respDesc).Debug("describe stream response")

	ch := make(chan *iteratorResult, len(respDesc.StreamDescription.Shards)) // buffered
	iterators := map[string]*string{}

	for _, shard := range respDesc.StreamDescription.Shards {
		go kh.asyncGetShardIterator(streamName, aws.StringValue(shard.ShardId), ch)
	}

	for range respDesc.StreamDescription.Shards {
		res := <-ch

		iterators[res.shardID] = res.iterator
	}

	return iterators, nil
}

func (kh *KinesisHelper) asyncGetShardIterator(streamName, shardID string, ch chan *iteratorResult) {
	kh.logger.WithField("shard", shardID).Debug("get shard iterator")

	respShard, err := kh.svc.GetShardIterator(&kinesis.GetShardIteratorInput{
		StreamName:        aws.String(streamName),
		ShardIteratorType: aws.String(kinesis.ShardIteratorTypeLatest),
		ShardId:           aws.String(shardID),
	})
	if err != nil {
		kh.logger.WithError(err).Fatal("get shard iterator failed")
	}

	ch <- &iteratorResult{shardID: shardID, iterator: respShard.ShardIterator}
}
