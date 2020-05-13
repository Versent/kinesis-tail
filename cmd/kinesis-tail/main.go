package main

import (
	"fmt"
	"io"
	"os"
	"runtime/trace"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/versent/kinesis-tail/pkg/rawdata"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/versent/kinesis-tail/pkg/ktail"
	"github.com/versent/kinesis-tail/pkg/logdata"
	"github.com/versent/kinesis-tail/pkg/sorter"
	"github.com/versent/kinesis-tail/pkg/streamer"
)

var (
	// Version program version which is updated via build flags
	version = "master"

	tracing       = kingpin.Flag("trace", "Enable trace mode.").Short('t').Bool()
	debug         = kingpin.Flag("debug", "Enable debug logging.").Short('d').Bool()
	region        = kingpin.Flag("region", "Configure the aws region.").Short('r').String()
	profile       = kingpin.Flag("profile", "Configure the aws profile.").Short('p').String()
	timestamp     = kingpin.Flag("timestamp", "Start time in epoch milliseconds.").Short('T').Int64()
	cwlogsCommand = kingpin.Command("cwlogs", "Process cloudwatch logs data from kinesis.")
	includes      = cwlogsCommand.Flag("include", "Include anything in log group names which match the supplied string.").Strings()
	excludes      = cwlogsCommand.Flag("exclude", "Exclude anything in log group names which match the supplied string.").Strings()
	cwlogsStream  = cwlogsCommand.Arg("stream", "Kinesis stream name.").Required().String()
	rawCommand    = kingpin.Command("raw", "Process raw data from kinesis.")
	rawStream     = rawCommand.Arg("stream", "Kinesis stream name.").Required().String()
	timeout       = rawCommand.Flag("timeout", "How long to capture raw data for before exiting in ms.").Default("3600000").Int64()
	count         = rawCommand.Flag("count", "How many records to capture raw data for before exiting.").Default("0").Int()

	logger = logrus.New()
)

func main() {
	kingpin.Version(version)
	subCommand := kingpin.Parse()

	if *tracing {
		f, err := os.Create("trace.out")
		if err != nil {
			logger.WithError(err).Fatal("failed to create trace file")
		}

		err = trace.Start(f)
		if err != nil {
			logger.WithError(err).Fatal("failed to start trace")
		}

		defer trace.Stop()
	}

	if *debug {
		// set debug globally
		logrus.SetLevel(logrus.DebugLevel)
		// set debug in the logger we already created
		logger.SetLevel(logrus.DebugLevel)
	}

	svc := newKinesis(region, profile)

	logger.WithField("timestamp", *timestamp).Debug("built kinesis service")

	switch subCommand {
	case "cwlogs":
		err := processLogData(svc, *cwlogsStream, *timestamp, *includes, *excludes)
		if err != nil {
			logger.WithError(err).Fatal("failed to process log data")
		}
	case "raw":
		err := processRawData(svc, *rawStream, *timeout, *timestamp, *count)
		if err != nil {
			logger.WithError(err).Fatal("failed to process log data")
		}
	}
}

func processLogData(svc kinesisiface.KinesisAPI, stream string, timestamp int64, includes, excludes []string) error {
	helper := ktail.New(svc, logger)

	iterators, err := helper.GetStreamIterators(stream, timestamp)
	if err != nil {
		return errors.Wrap(err, "get iterators failed")
	}

	kstream := streamer.New(svc, iterators, 5000, logger)
	ch := kstream.StartGetRecords()

	messageSorter := sorter.New(os.Stdout, len(iterators), formatLogsMsg)

	for result := range ch {
		logger.WithField("count", len(result.Records)).WithField("shard", result.Shard).Debug("received records")

		if result.Err != nil {
			return errors.Wrap(result.Err, "get records failed")
		}

		msgResults := []*ktail.LogMessage{}

		for _, rec := range result.Records {
			msgs, err := logdata.UncompressLogs(includes, excludes, rec.ApproximateArrivalTimestamp, rec.Data)
			if err != nil {
				return errors.Wrap(err, "parse log records failed")
			}

			msgResults = append(msgResults, msgs...)
		}

		messageSorter.PushBatch(msgResults)
	}

	return nil
}

func processRawData(svc kinesisiface.KinesisAPI, stream string, timeout, timestamp int64, count int) error {
	helper := ktail.New(svc, logger)

	iterators, err := helper.GetStreamIterators(stream, timestamp)
	if err != nil {
		return errors.Wrap(err, "get iterators failed")
	}

	kstream := streamer.New(svc, iterators, 5000, logger)
	ch := kstream.StartGetRecords()

	messageSorter := sorter.New(os.Stdout, len(iterators), formatRawMsg)

	timer1 := time.NewTimer(time.Duration(timeout) * time.Millisecond)

	if count > 0 {
		logger.WithField("count", count).Debug("waiting for records")
	}

	var recordCount int

LOOP:
	for {
		select {
		case result := <-ch:
			if result.Err != nil {
				return errors.Wrap(result.Err, "get records failed")
			}

			logger.WithFields(logrus.Fields{
				"count": len(result.Records),
				"total": recordCount,
				"shard": result.Shard,
			}).Debug("received records")

			msgResults := []*ktail.LogMessage{}

			for _, rec := range result.Records {
				msg := rawdata.DecodeRawData(rec.ApproximateArrivalTimestamp, rec.Data)
				msgResults = append(msgResults, msg)
			}

			messageSorter.PushBatch(msgResults)

			recordCount += len(result.Records)

			if count != 0 {
				if recordCount >= count {
					messageSorter.Flush()

					logger.WithField("recordCount", recordCount).Info("reached count exit")
					break LOOP
				}
			}

		case <-timer1.C:
			logger.Info("timer expired exit")
			break LOOP
		}
	}

	return nil
}

func formatRawMsg(wr io.Writer, msg *ktail.LogMessage) {
	_, err := fmt.Fprintln(wr, msg.Message)
	if err != nil {
		logger.WithError(err).Fatal("failed to create trace file")
	}
}

func formatLogsMsg(wr io.Writer, msg *ktail.LogMessage) {
	c := color.New(color.FgBlue)
	_, err := fmt.Fprintf(wr, "%s %s\n", c.Sprintf("[%s %s]", msg.Timestamp, msg.LogGroup), msg.Message)
	if err != nil {
		logger.WithError(err).Fatal("failed to create trace file")
	}
}

func newKinesis(region, profile *string) kinesisiface.KinesisAPI {
	sess := session.Must(session.NewSession())

	cfg := aws.NewConfig()

	if aws.StringValue(region) != "" {
		cfg = cfg.WithRegion(*region)
	}

	if aws.StringValue(profile) != "" {
		cfg = cfg.WithCredentials(credentials.NewSharedCredentials("", *profile))
	}

	return kinesis.New(sess, cfg)
}
