package ses

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/go-mailx/mailx"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

type mailerAdapter struct {
	client *ses.Client
}

func (a *mailerAdapter) NewMail(ctx context.Context) (mailx.MailInstance, error) {
	return &messageAdapter{
		client: a.client,
		msg: ses.SendEmailInput{
			Destination: &types.Destination{},
			Message:     &types.Message{},
		},
	}, nil
}

func New(config aws.Config) *mailerAdapter {
	return &mailerAdapter{
		client: ses.NewFromConfig(config),
	}
}

func NewFromContext(ctx context.Context) (*mailerAdapter, error) {
	if cfg, err := config.LoadDefaultConfig(ctx); err != nil {
		return nil, err
	} else {
		otelaws.AppendMiddlewares(&cfg.APIOptions)
		return New(cfg), nil
	}
}

type messageAdapter struct {
	client *ses.Client
	msg    ses.SendEmailInput
}

func (m *messageAdapter) Bcc(bccs []string) error {
	m.msg.Destination.BccAddresses = append(m.msg.Destination.BccAddresses, bccs...)
	return nil
}

func (m *messageAdapter) From(from string) error {
	m.msg.Source = &from
	return nil
}

func (m *messageAdapter) HtmlBody(body string) error {
	if m.msg.Message.Body == nil {
		m.msg.Message.Body = &types.Body{}
	}
	m.msg.Message.Body.Html = &types.Content{Data: &body}
	return nil
}

func (m *messageAdapter) ReplyTo(replyTo string) error {
	return m.ReplyTo(replyTo)
}

func (m *messageAdapter) Send(ctx context.Context) error {
	_, err := m.client.SendEmail(ctx, &m.msg)
	return err
}

func (m *messageAdapter) Subject(sub string) error {
	m.msg.Message.Subject = &types.Content{Data: &sub}
	return nil
}

func (m *messageAdapter) TextBody(body string) error {
	if m.msg.Message.Body == nil {
		m.msg.Message.Body = &types.Body{}
	}
	m.msg.Message.Body.Text = &types.Content{Data: &body}
	return nil
}

func (m *messageAdapter) To(to []string) error {
	m.msg.Destination.ToAddresses = append(m.msg.Destination.ToAddresses, to...)
	return nil
}
