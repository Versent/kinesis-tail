# kinesis-tail

Tool which provides tail for [Kinesis](https://aws.amazon.com/kinesis/streams/), it allows you to use one of two processors for the data returned, firstly one which decompresses and parses [CloudWatch Logs](http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/WhatIsCloudWatchLogs.html) data, and secondly one which just returns the raw data.

# background

The cloudwatch logs reader is designed to work with a common pattern used at Versent for log distribution and storage.

For more information on the setup for `cwlogs` sub command to function it assumes the logs are gzipped batches of log JSON records in Kinesis see [Real-time Processing of Log Data with Subscriptions](http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CreateDestination.html)

# usage

```
$ kinesis-tail --help-long
usage: kinesis-tail [<flags>] <command> [<args> ...]

Flags:
      --help           Show context-sensitive help (also try --help-long and --help-man).
  -t, --trace          Enable trace mode.
  -r, --region=REGION  Configure the aws region.
      --version        Show application version.

Commands:
  help [<command>...]
    Show help.


  cwlogs [<flags>] <stream>
    Process cloudwatch logs data from kinesis.

    --include=INCLUDE ...  Include anything in log group names which match the supplied string.
    --exclude=EXCLUDE ...  Exclude anything in log group names which match the supplied string.

  raw <stream>
    Process raw data from kinesis.

```

# license

This code is released under MIT License.

