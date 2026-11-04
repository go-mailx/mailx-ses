# mailx-ses

[![Go Reference](https://pkg.go.dev/badge/github.com/go-mailx/mailx-ses.svg)](https://pkg.go.dev/github.com/go-mailx/mailx-ses)

`github.com/go-mailx/mailx-ses` is an AWS SES adapter for [`github.com/go-mailx/mailx`](https://github.com/go-mailx/mailx). It implements the `mailx.MailerAdapter` interface using the AWS SDK v2 and includes OpenTelemetry instrumentation via `otelaws`.

## Install

```sh
go get github.com/go-mailx/mailx-ses
```

## Usage

See the [ses example](https://github.com/go-mailx/mailx/tree/main/examples/ses) in the main repo.

### Inline attachments

The adapter supports inline (embedded) attachments added with `mailx.Inline`, e.g. an image that the HTML body references as `<img src="cid:logo">`:

```go
mailer.Send(ctx,
	mailx.To("to@example.com"),
	mailx.HtmlBody(`<img src="cid:logo">`),
	mailx.Inline("logo", "image/png", logo),
)
```

Mails are sent with the SES v2 `SendEmail` API, which requires the IAM action `ses:SendEmail`.

## Config

See [`pkg.go.dev/github.com/go-mailx/mailx-ses`](https://pkg.go.dev/github.com/go-mailx/mailx-ses) for all configuration options.
