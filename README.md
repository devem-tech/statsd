
# StatsD Client for Go

A lightweight, efficient StatsD client library in Go that allows you to send metrics to a StatsD server. It supports configurable buffering, automatic flushing, and custom tags. Ideal for applications that need low-latency, low-overhead metric tracking.

## Features

- **High Performance**: Optimized for high loads, with minimal latency and efficient buffering.
- **Buffered Metrics**: Efficiently buffers metrics to reduce network overhead.
- **Configurable Options**: Set custom buffer sizes, flush intervals, and error handling.
- **Automatic Flushing**: Periodic flushing of buffered metrics.
- **Customizable Tags and Prefixes**: Tag metrics with key-value pairs and add prefixes for easier organization.

## Installation

Add this library to your Go module using:

```bash
go get github.com/devem-tech/statsd
```

## Usage

### 1. Basic Usage

Here’s a simple example of how to create a StatsD client, send metrics, and configure options.

```go
package main

import (
    "log"
    "time"
    "github.com/devem-tech/statsd"
)

func main() {
    client, err := statsd.New(
        statsd.Host("localhost"),
        statsd.Port(8125),
        statsd.Prefix("app"),
        statsd.FlushInterval(2 * time.Second),
        statsd.ErrorHandler(func(err error) {
            log.Println("StatsD error:", err)
        }),
    )
    if err != nil {
        log.Fatal("Failed to create StatsD client:", err)
    }
    defer client.Close()

    client.Increment("user.signup")
    client.Gauge("system.load", 1.5)
    client.Timing("db.query_time", 250*time.Millisecond)
}
```

### 2. Configuring Options

You can configure the client with options to match your requirements. Here’s a breakdown of some useful options:

- **Host**: Specify the StatsD server hostname or IP.
- **Port**: Define the UDP port for the StatsD server.
- **MaxBufferSize**: Set the maximum buffer size in bytes before triggering a flush.
- **FlushInterval**: Define how often the buffer should automatically flush.
- **ErrorHandler**: Provide a custom function for handling errors.
- **Prefix**: Add a prefix to all metric names.
- **Tags**: Define global tags to be added to every metric.

Example:

```go
client, err := statsd.New(
    statsd.Host("localhost"),
    statsd.Port(8125),
    statsd.MaxBufferSize(1024),
    statsd.FlushInterval(500 * time.Millisecond),
    statsd.Prefix("app"),
    statsd.Tags([]statsd.Tag{
        {Key: "env", Value: "production"},
    }),
)
```

### 3. Metric Types

This library supports the following metric types:

- **Count**: Increments a counter metric.

  ```go
  client.Count("user.signup", 1)
  ```

- **Increment**: Increments a counter by 1.

  ```go
  client.Increment("page.views")
  ```

- **Gauge**: Records a gauge value.

  ```go
  client.Gauge("cpu.usage", 72.5)
  ```

- **Timing**: Records a timing value in milliseconds.

  ```go
  client.Timing("response.time", 350*time.Millisecond)
  ```

- **Timer**: Starts a timer for measuring duration and sends the result on completion.

  ```go
  stopTimer := client.Timer("db.query_time")
  defer stopTimer() // Automatically records duration when done
  ```

### 4. Closing the Client

Always close the client to ensure all metrics are flushed and resources are released.

```go
client.Close()
```

## Advanced Usage

### Custom Error Handling

You can provide an error handler to log or take action on any errors that occur during metric transmission.

```go
client, err := statsd.New(
    statsd.ErrorHandler(func(err error) {
        log.Printf("StatsD error: %v", err)
    }),
)
```

### Adding Default Tags

To add default tags that are sent with every metric:

```go
client, err := statsd.New(
    statsd.Tags([]statsd.Tag{
        {Key: "service", Value: "api"},
        {Key: "env", Value: "prod"},
    }),
)
```

## Contributing

We welcome contributions to improve this library.  
Please submit issues and pull requests on the GitHub repository.

## License

This library is licensed under the MIT License.
