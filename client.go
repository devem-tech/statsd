package statsd

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	defaultPort          = 8125
	defaultMaxBufferSize = 512
	defaultFlushInterval = 100 * time.Millisecond
)

const bufferCapFactor = 2

// options represent the client configuration.
type options struct {
	host          string
	port          int
	maxBufferSize int
	flushInterval time.Duration
	errorHandler  func(error)
	prefix        string
	tags          []Tag
}

// Tag represents a key-value pair used for tagging metrics.
type Tag struct {
	Key   string
	Value string
}

// Client represents a StatsD client.
type Client struct {
	conn          net.Conn
	buffer        []byte
	bufferLock    sync.Mutex
	maxBufferSize int
	flushInterval time.Duration
	flushChan     chan struct{}
	quitChan      chan struct{}
	wg            sync.WaitGroup
	errorHandler  func(error)
	prefix        []byte
	tags          []byte
}

// New returns a new Client.
func New(opts ...Option) (*Client, error) {
	o := &options{
		host:          "",
		port:          defaultPort,
		maxBufferSize: defaultMaxBufferSize,
		flushInterval: defaultFlushInterval,
		errorHandler:  nil,
		prefix:        "",
		tags:          nil,
	}

	for _, opt := range opts {
		opt(o)
	}

	conn, err := net.Dial("udp", o.host+":"+strconv.Itoa(o.port))
	if err != nil {
		return nil, fmt.Errorf("statsd: %w", err)
	}

	client := &Client{
		conn:          conn,
		buffer:        make([]byte, 0, o.maxBufferSize*bufferCapFactor),
		bufferLock:    sync.Mutex{},
		maxBufferSize: o.maxBufferSize,
		flushInterval: o.flushInterval,
		flushChan:     make(chan struct{}, 1), // Buffer by 1 to prevent locks
		quitChan:      make(chan struct{}),
		wg:            sync.WaitGroup{},
		errorHandler:  o.errorHandler,
		prefix:        []byte(o.prefix),
		tags:          make([]byte, 0, o.maxBufferSize),
	}

	client.serializeTagsTo(client.tags, o.tags)
	client.startBackgroundFlusher()

	return client, nil
}

// startBackgroundFlusher starts the background flusher to send metrics regularly.
func (c *Client) startBackgroundFlusher() {
	c.wg.Add(1)

	go func() {
		defer c.wg.Done()

		ticker := time.NewTicker(c.flushInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Request flushing through the channel
				c.requestFlush()
			case <-c.flushChan:
				// When the channel receives a signal, we flush the metrics
				c.flushMetrics()
			case <-c.quitChan:
				// Closing, final flush
				c.flushMetrics()

				return
			}
		}
	}()
}

// requestFlush sends a signal for flushing through the channel.
func (c *Client) requestFlush() {
	// Do not block if the flush is already
	// in process (there is already a signal
	// in the channel).
	select {
	case c.flushChan <- struct{}{}:
	default:
	}
}

// send adds the metric to the buffer instead of sending it immediately.
func (c *Client) send(key, value, mt string, tags ...Tag) {
	c.bufferLock.Lock()
	defer c.bufferLock.Unlock()

	c.buffer = append(c.buffer, c.prefix...)
	c.buffer = append(c.buffer, key...)
	c.buffer = append(c.buffer, ':')
	c.buffer = append(c.buffer, value...)
	c.buffer = append(c.buffer, c.tags...)
	c.serializeTagsTo(c.buffer, tags)
	c.buffer = append(c.buffer, '|')
	c.buffer = append(c.buffer, mt...)
	c.buffer = append(c.buffer, '\n')

	// If the buffer is full, request flushing
	if len(c.buffer) >= c.maxBufferSize {
		c.requestFlush()
	}
}

// flushMetrics sends all metrics from the buffer to StatsD.
func (c *Client) flushMetrics() {
	c.bufferLock.Lock()

	n := len(c.buffer)
	if n == 0 {
		c.bufferLock.Unlock()

		return
	}

	data := c.buffer[:n-1]
	c.buffer = c.buffer[:0]
	c.bufferLock.Unlock()

	_, err := c.conn.Write(data)
	if err != nil && c.errorHandler != nil {
		c.errorHandler(err)
	}
}

// serializeTagsTo serializes the tags into a byte slice.
func (c *Client) serializeTagsTo(buffer []byte, tags []Tag) {
	for _, tag := range tags {
		buffer = append(buffer, ';')
		buffer = append(buffer, tag.Key...)
		buffer = append(buffer, '=')
		buffer = append(buffer, tag.Value...)
	}
}

// Count sends a counter.
func (c *Client) Count(key string, value int64, tags ...Tag) {
	if value == 0 {
		return
	}

	c.send(key, strconv.FormatInt(value, 10), "c", tags...)
}

// Increment increases a counter by 1.
func (c *Client) Increment(key string, tags ...Tag) {
	c.Count(key, 1, tags...)
}

// Gauge sends a gauge.
func (c *Client) Gauge(key string, value float64, tags ...Tag) {
	c.send(key, strconv.FormatFloat(value, 'f', -1, 64), "g", tags...)
}

// Timing sends a timer.
func (c *Client) Timing(key string, duration time.Duration, tags ...Tag) {
	c.send(key, strconv.FormatInt(duration.Milliseconds(), 10), "ms", tags...)
}

// Timer starts timing and sends the metric via defer.
func (c *Client) Timer(key string, tags ...Tag) func() {
	start := time.Now()

	return func() {
		c.Timing(key, time.Since(start), tags...)
	}
}

// Close closes the connection with StatsD and flushes the remaining metrics.
func (c *Client) Close() {
	close(c.quitChan)

	c.wg.Wait() // Wait for background tasks to finish

	err := c.conn.Close()
	if err != nil && c.errorHandler != nil {
		c.errorHandler(err)
	}
}
