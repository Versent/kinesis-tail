# kinesis-tail

Tool which provides tail for [Kinesis](https://aws.amazon.com/kinesis/streams/), it allows you to use one of two processors for the data returned, firstly one which decompresses and parses [CloudWatch Logs](http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/WhatIsCloudWatchLogs.html) data, and secondly one which just returns the raw data.

# background

This cloudwatch logs reader works with a pattern used at Versent for log distribution and storage.

For more information on the setup for `cwlogs` sub command to function it assumes the logs are gzipped batches of log JSON records in Kinesis see [Real-time Processing of Log Data with Subscriptions](http://docs.aws.amazon.com/AmazonCloudWatch/latest/logs/CreateDestination.html)

# installation

You can download `kinesis-tail` from [Releases](https://github.com/Versent/kinesis-tail/releases) or install it using npm.

# usage

```
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

  raw [<flags>] <stream>
    Process raw data from kinesis.

    --timeout=3600000  How long to capture raw data for before exiting in ms.
    --count=0          How many records to capture raw data for before exiting.


```

# example

List the kinesis streams in your account.

```
aws kinesis list-streams
```

To tail one of these streams and exit once you have captured 20 records.

```
kinesis-tail raw dev-1-stream --count 20
```

To tail one of these streams and exit after 30 seconds, and write the data to a file.

```
kinesis-tail raw dev-1-stream --timeout 30000 | tee data.log
```

# license

This code is released under MIT License.

