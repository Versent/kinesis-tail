package main

import (
	"fmt"
	"io"
	"os"
	"runtime/trace"

	"github.com/alecthomas/kingpin"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/versent/kinesis-tail/pkg/ktail"
	"github.com/versent/kinesis-tail/pkg/logdata"
	"github.com/versent/kinesis-tail/pkg/sorter"
	"github.com/versent/kinesis-tail/pkg/streamer"
)

var (
	// Version program version which is updated via build flags
	Version = "1.0.0"

	tracing  = kingpin.Flag("trace", "Enable trace mode.").Short('t').Bool()
	region   = kingpin.Flag("region", "Configure the aws region.").Short('r').String()
	stream   = kingpin.Arg("stream", "Kinesis stream name.").Required().String()
	includes = kingpin.Flag("include", "Include anything in log group names which match the supplied string.").Strings()
	excludes = kingpin.Flag("exclude", "Exclude anything in log group names which match the supplied string.").Strings()
	format   = kingpin.Flag("format", "Formatter used to output messages.").Default("cwlogs").Enum("cwlogs", "raw")
)

func main() {
	kingpin.Version(Version)
	kingpin.Parse()

	logger := logrus.New()

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

	helper := ktail.New(svc, logger)

	iterators, err := helper.GetStreamIterators(*stream)
	if err != nil {
		logger.WithError(err).Fatal("get iterators failed")
	}

	var messageSorter *sorter.MessageSorter

	switch *format {
	case "raw":
		messageSorter = sorter.New(os.Stdout, len(iterators), formatRawMsg)
	case "cwlogs":
		messageSorter = sorter.New(os.Stdout, len(iterators), formatLogsMsg)
	}

	kstream := streamer.New(svc, iterators, 5000)
	ch := kstream.StartGetRecords()

	for result := range ch {

		if result.Err != nil {
			logger.WithError(result.Err).Fatal("get records failed")
		}

		msgResults := []*ktail.LogMessage{}

		for _, rec := range result.Records {
			msgs, err := logdata.UncompressLogs(*includes, *excludes, rec.ApproximateArrivalTimestamp, rec.Data)
			if err != nil {
				logger.WithError(err).Fatal("parse log records failed")
			}

			msgResults = append(msgResults, msgs...)
		}

		messageSorter.PushBatch(msgResults)
	}
}

func formatRawMsg(wr io.Writer, msg *ktail.LogMessage) {
	fmt.Fprintln(wr, msg.Message)
}

func formatLogsMsg(wr io.Writer, msg *ktail.LogMessage) {
	c := color.New(color.FgBlue)
	fmt.Fprintf(wr, "%s %s\n", c.Sprintf("[%s %s]", msg.Timestamp, msg.LogGroup), msg.Message)
}
