# mailx-ses

[![Go Reference](https://pkg.go.dev/badge/github.com/go-mailx/mailx-ses.svg)](https://pkg.go.dev/github.com/go-mailx/mailx-ses)

`github.com/go-mailx/mailx-ses` is an AWS SES adapter for [`github.com/go-mailx/mailx`](../mailx). It implements the `mailx.MailerAdapter` interface using the AWS SDK v2 and includes OpenTelemetry instrumentation via `otelaws`.

## Install

```sh
go get github.com/go-mailx/mailx-ses
```

## Usage

See the [ses example](https://github.com/go-mailx/mailx/tree/main/examples/ses) in the main repo.

## Config

See [`pkg.go.dev/github.com/go-mailx/mailx-ses`](https://pkg.go.dev/github.com/go-mailx/mailx-ses) for all configuration options.
