package ses

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/go-mailx/mailx"
)

// captureSendEmailInput returns an aws.Config whose clients record the
// SendEmailInput of each call and return before any request is sent.
func captureSendEmailInput(captured **sesv2.SendEmailInput) aws.Config {
	return aws.Config{
		Region: "eu-central-1",
		APIOptions: []func(*middleware.Stack) error{
			func(stack *middleware.Stack) error {
				return stack.Initialize.Add(middleware.InitializeMiddlewareFunc("captureInput",
					func(ctx context.Context, in middleware.InitializeInput, next middleware.InitializeHandler) (middleware.InitializeOutput, middleware.Metadata, error) {
						*captured = in.Parameters.(*sesv2.SendEmailInput)
						return middleware.InitializeOutput{Result: &sesv2.SendEmailOutput{}}, middleware.Metadata{}, nil
					}), middleware.Before)
			},
		},
	}
}

func TestSendEmailInput(t *testing.T) {
	var got *sesv2.SendEmailInput
	mailer := &mailx.Mailer{
		MailerAdapter: New(captureSendEmailInput(&got)),
		Config:        &mailx.MailerConfig{FromAddressSrc: []mailx.FromAddressFunc{mailx.StaticFromAddress("from@example.com")}},
	}
	logo := []byte("\x89PNG\r\n\x1a\n")
	err := mailer.Send(context.Background(),
		mailx.To("to@example.com"),
		mailx.Bcc("bcc1@example.com", "bcc2@example.com"),
		mailx.ReplyTo("reply@example.com"),
		mailx.Subject("Grüße"),
		mailx.HtmlBody(`<p>Hi</p><img src="cid:logo">`),
		mailx.TextBody("Hi"),
		mailx.Inline("logo", "image/png", logo),
	)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	want := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String("from@example.com"),
		Destination: &types.Destination{
			ToAddresses:  []string{"to@example.com"},
			BccAddresses: []string{"bcc1@example.com", "bcc2@example.com"},
		},
		ReplyToAddresses: []string{"reply@example.com"},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String("Grüße"), Charset: aws.String("UTF-8")},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(`<p>Hi</p><img src="cid:logo">`), Charset: aws.String("UTF-8")},
					Text: &types.Content{Data: aws.String("Hi"), Charset: aws.String("UTF-8")},
				},
				Attachments: []types.Attachment{{
					FileName:                aws.String("logo.png"),
					RawContent:              logo,
					ContentId:               aws.String("logo"),
					ContentType:             aws.String("image/png"),
					ContentDisposition:      types.AttachmentContentDispositionInline,
					ContentTransferEncoding: types.AttachmentContentTransferEncodingBase64,
				}},
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.MarshalIndent(got, "", "  ")
		wantJSON, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("SendEmailInput mismatch\ngot:  %s\nwant: %s", gotJSON, wantJSON)
	}
}

func TestInlineFileName(t *testing.T) {
	tests := []struct {
		contentID, contentType, want string
	}{
		{"logo", "image/png", "logo.png"},
		{"photo", "image/jpeg", "photo.jpeg"},
		{"icon", "image/svg+xml", "icon.svg"},
		{"logo.png", "image/png", "logo.png"},
		{"photo.JPG", "image/jpeg", "photo.JPG"},
		{"logo@example.com", "image/gif", "logo@example.com.gif"},
		{"a/b", "image/png", "a_b.png"},
		{"blob", "application/x-unknown-type", "blob"},
	}
	for _, tt := range tests {
		if got := inlineFileName(mailx.InlineAttachment{ContentID: tt.contentID, ContentType: tt.contentType}); got != tt.want {
			t.Errorf("inlineFileName(%q, %q) = %q, want %q", tt.contentID, tt.contentType, got, tt.want)
		}
	}
}
