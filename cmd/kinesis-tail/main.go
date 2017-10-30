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
	Version = "1.0.0"

	tracing       = kingpin.Flag("trace", "Enable trace mode.").Short('t').Bool()
	region        = kingpin.Flag("region", "Configure the aws region.").Short('r').String()
	cwlogsCommand = kingpin.Command("cwlogs", "Process cloudwatch logs data from kinesis.")
	includes      = cwlogsCommand.Flag("include", "Include anything in log group names which match the supplied string.").Strings()
	excludes      = cwlogsCommand.Flag("exclude", "Exclude anything in log group names which match the supplied string.").Strings()
	cwlogsStream  = cwlogsCommand.Arg("stream", "Kinesis stream name.").Required().String()
	rawCommand    = kingpin.Command("raw", "Process raw data from kinesis.")
	rawStream     = rawCommand.Arg("stream", "Kinesis stream name.").Required().String()
	timeout       = rawCommand.Flag("timeout", "How long to capture raw data for before exiting in ms.").Default("3600000").Int64()

	logger = logrus.New()
)

func main() {
	kingpin.Version(Version)
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

	sess := session.Must(session.NewSession())

	var svc kinesisiface.KinesisAPI

	if *region == "" {
		svc = kinesis.New(sess)
	} else {
		// Create a Kinesis client with additional configuration
		svc = kinesis.New(sess, aws.NewConfig().WithRegion(*region))
	}

	logger.Debug("built kinesis service")

	switch subCommand {
	case "cwlogs":
		err := processLogData(svc, *cwlogsStream, *includes, *excludes)
		if err != nil {
			logger.WithError(err).Fatal("failed to process log data")
		}
	case "raw":
		err := processRawData(svc, *rawStream, *timeout)
		if err != nil {
			logger.WithError(err).Fatal("failed to process log data")
		}
	}

}

func processLogData(svc kinesisiface.KinesisAPI, stream string, includes []string, excludes []string) error {

	helper := ktail.New(svc, logger)

	iterators, err := helper.GetStreamIterators(stream)
	if err != nil {
		return errors.Wrap(err, "get iterators failed")
	}

	kstream := streamer.New(svc, iterators, 5000)
	ch := kstream.StartGetRecords()

	messageSorter := sorter.New(os.Stdout, len(iterators), formatLogsMsg)

	for result := range ch {

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

func processRawData(svc kinesisiface.KinesisAPI, stream string, timeout int64) error {

	helper := ktail.New(svc, logger)

	iterators, err := helper.GetStreamIterators(stream)
	if err != nil {
		return errors.Wrap(err, "get iterators failed")
	}

	kstream := streamer.New(svc, iterators, 5000)
	ch := kstream.StartGetRecords()

	messageSorter := sorter.New(os.Stdout, len(iterators), formatRawMsg)

	timer1 := time.NewTimer(time.Duration(timeout) * time.Millisecond)

LOOP:
	for {

		select {
		case result := <-ch:
			if result.Err != nil {
				return errors.Wrap(result.Err, "get records failed")
			}

			logger.WithField("count", len(result.Records)).WithField("shard", result.Shard).Info("received records")

			msgResults := []*ktail.LogMessage{}

			for _, rec := range result.Records {
				msg := rawdata.DecodeRawData(rec.ApproximateArrivalTimestamp, rec.Data)
				msgResults = append(msgResults, msg)
			}

			messageSorter.PushBatch(msgResults)
		case <-timer1.C:
			logger.Info("timer expired exit")
			break LOOP
		}

	}

	return nil
}

func formatRawMsg(wr io.Writer, msg *ktail.LogMessage) {
	fmt.Fprintln(wr, msg.Message)
}

func formatLogsMsg(wr io.Writer, msg *ktail.LogMessage) {
	c := color.New(color.FgBlue)
	fmt.Fprintf(wr, "%s %s\n", c.Sprintf("[%s %s]", msg.Timestamp, msg.LogGroup), msg.Message)
}
