package ktail

// LogEntry matches the cloudwatch log entry structure
type LogEntry struct {
	ID        string `json:"id,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	Message   string `json:"message,omitempty"`
}

// LogBatch matches the cloudwatch logs batch structure
type LogBatch struct {
	MessageType         string      `json:"messageType,omitempty"`
	Owner               string      `json:"owner,omitempty"`
	LogGroup            string      `json:"logGroup,omitempty"`
	LogStream           string      `json:"logStream,omitempty"`
	SubscriptionFilters []string    `json:"subscriptionFilters,omitempty"`
	LogEvents           []*LogEntry `json:"logEvents,omitempty"`
}

// LogMessage log message after decompression and parsing
type LogMessage struct {
	LogGroup  string
	Message   string
	Timestamp string
}

// ByTimestamp used to sort log messages
type ByTimestamp []*LogMessage

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool { return a[i].Timestamp < a[j].Timestamp }
