package ses

import (
	"context"
	"mime"
	"path"
	"slices"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/go-mailx/mailx"
	"go.opentelemetry.io/contrib/instrumentation/github.com/aws/aws-sdk-go-v2/otelaws"
)

const charset = "UTF-8"

type mailerAdapter struct {
	client *sesv2.Client
}

func (a *mailerAdapter) NewMail(ctx context.Context) (mailx.MailInstance, error) {
	return &messageAdapter{
		client: a.client,
		msg: sesv2.SendEmailInput{
			Destination: &types.Destination{},
			Content:     &types.EmailContent{Simple: &types.Message{}},
		},
	}, nil
}

func New(config aws.Config) *mailerAdapter {
	return &mailerAdapter{
		client: sesv2.NewFromConfig(config),
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
	client *sesv2.Client
	msg    sesv2.SendEmailInput
}

var _ mailx.InlineAttachmentAdder = (*messageAdapter)(nil)

func (m *messageAdapter) Bcc(bccs []string) error {
	m.msg.Destination.BccAddresses = append(m.msg.Destination.BccAddresses, bccs...)
	return nil
}

func (m *messageAdapter) From(from string) error {
	m.msg.FromEmailAddress = &from
	return nil
}

func (m *messageAdapter) HtmlBody(body string) error {
	m.body().Html = &types.Content{Data: &body, Charset: aws.String(charset)}
	return nil
}

func (m *messageAdapter) ReplyTo(replyTo string) error {
	m.msg.ReplyToAddresses = append(m.msg.ReplyToAddresses, replyTo)
	return nil
}

func (m *messageAdapter) Send(ctx context.Context) error {
	_, err := m.client.SendEmail(ctx, &m.msg)
	return err
}

func (m *messageAdapter) Subject(sub string) error {
	m.msg.Content.Simple.Subject = &types.Content{Data: &sub, Charset: aws.String(charset)}
	return nil
}

func (m *messageAdapter) TextBody(body string) error {
	m.body().Text = &types.Content{Data: &body, Charset: aws.String(charset)}
	return nil
}

func (m *messageAdapter) To(to []string) error {
	m.msg.Destination.ToAddresses = append(m.msg.Destination.ToAddresses, to...)
	return nil
}

func (m *messageAdapter) AddInlineAttachment(a mailx.InlineAttachment) error {
	m.msg.Content.Simple.Attachments = append(m.msg.Content.Simple.Attachments, types.Attachment{
		FileName:                aws.String(inlineFileName(a)),
		RawContent:              a.Data,
		ContentId:               aws.String(a.ContentID),
		ContentType:             aws.String(a.ContentType),
		ContentDisposition:      types.AttachmentContentDispositionInline,
		ContentTransferEncoding: types.AttachmentContentTransferEncodingBase64,
	})
	return nil
}

func (m *messageAdapter) body() *types.Body {
	if m.msg.Content.Simple.Body == nil {
		m.msg.Content.Simple.Body = &types.Body{}
	}
	return m.msg.Content.Simple.Body
}

// inlineFileName derives a file name such as "logo.png" from the Content-ID
// and content type, because SES requires a file name for every attachment.
func inlineFileName(a mailx.InlineAttachment) string {
	name := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' {
			return '_'
		}
		return r
	}, a.ContentID)
	exts, _ := mime.ExtensionsByType(a.ContentType)
	if len(exts) == 0 || slices.Contains(exts, strings.ToLower(path.Ext(name))) {
		return name
	}
	// The extension list depends on the system's mime.types, so prefer the one
	// matching the subtype (".jpeg" for image/jpeg) over the alphabetically first.
	if mediaType, _, err := mime.ParseMediaType(a.ContentType); err == nil {
		if _, subtype, ok := strings.Cut(mediaType, "/"); ok && slices.Contains(exts, "."+subtype) {
			return name + "." + subtype
		}
	}
	return name + exts[0]
}
