package models

import (
	"bytes"
	"net/mail"
	"net/url"
	"path"
	"reflect"
	"text/template"
)

// TemplateContext is an interface that allows both campaigns and email
// requests to have a PhishingTemplateContext generated for them.
type TemplateContext interface {
	getFromAddress() string
	getBaseURL() string
}

// PhishingTemplateContext is the context that is sent to any template, such
// as the email or landing page content.
type PhishingTemplateContext struct {
	From          string
	URL           string
	Tracker       string
	TrackingURL   string
	RId           string
	BaseURL       string
	AttachmentURL string
	BaseRecipient
}

// NewPhishingTemplateContext returns a populated PhishingTemplateContext,
// parsing the correct fields from the provided TemplateContext and recipient.
func NewPhishingTemplateContext(ctx TemplateContext, r BaseRecipient, rid string) (PhishingTemplateContext, error) {
	f, err := mail.ParseAddress(ctx.getFromAddress())
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	templateURL, err := ExecuteTemplate(ctx.getBaseURL(), r)
	if err != nil {
		return PhishingTemplateContext{}, err
	}

	// For the base URL, we'll reset the the path and the query
	// This will create a URL in the form of http://example.com
	baseURL, err := url.Parse(templateURL)
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	baseURL.Path = ""
	baseURL.RawQuery = ""

	phishURL, _ := url.Parse(templateURL)
	q := phishURL.Query()
	q.Set(RecipientParameter, rid)
	phishURL.RawQuery = q.Encode()

	trackingURL, _ := url.Parse(templateURL)
	trackingURL.Path = path.Join(trackingURL.Path, "/track")
	trackingURL.RawQuery = q.Encode()

	attachmentURL, _ := url.Parse(templateURL)
	attachmentURL.Path = path.Join(trackingURL.Path, "attachment")
	attachmentURL.RawQuery = q.Encode()

	return PhishingTemplateContext{
		BaseRecipient: r,
		BaseURL:       baseURL.String(),
		URL:           phishURL.String(),
		TrackingURL:   trackingURL.String(),
		AttachmentURL: attachmentURL.String(),
		Tracker:       "<img alt='' style='display: none' src='" + trackingURL.String() + "'/>",
		From:          fn,
		RId:           rid,
	}, nil
}

// ExecuteTemplate creates a templated string based on the provided
// template body and data.
func ExecuteTemplateFull(text string, data interface{}, isAttachment bool) (string, error) {
	buff := bytes.Buffer{}
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return buff.String(), err
	}

	if isAttachment {
		var ptx = data.(PhishingTemplateContext)
		orig := &ptx
		ptx2 := *orig
		ptx2.TrackingURL = ptx2.AttachmentURL
		err = tmpl.Execute(&buff, ptx2)
	} else {
		err = tmpl.Execute(&buff, data)
	}

	return buff.String(), err
}
func ExecuteTemplate(text string, data interface{}) (string, error) {
	return ExecuteTemplateFull(text, data, false)
}

// ValidationContext is used for validating templates and pages
type ValidationContext struct {
	FromAddress string
	BaseURL     string
}

func (vc ValidationContext) getFromAddress() string {
	return vc.FromAddress
}

func (vc ValidationContext) getBaseURL() string {
	return vc.BaseURL
}

// ValidateTemplate ensures that the provided text in the page or template
// uses the supported template variables correctly.
func ValidateTemplate(text string) error {
	vc := ValidationContext{
		FromAddress: "foo@bar.com",
		BaseURL:     "http://example.com",
	}
	td := Result{
		BaseRecipient: BaseRecipient{
			Email:     "foo@bar.com",
			FirstName: "Foo",
			LastName:  "Bar",
			Position:  "Test",
		},
		RId: "123456",
	}
	ptx, err := NewPhishingTemplateContext(vc, td.BaseRecipient, td.RId)
	if err != nil {
		return err
	}
	_, err = ExecuteTemplate(text, ptx)
	if err != nil {
		return err
	}
	return nil
}

func Clone(oldObj interface{}) interface{} {
	newObj := reflect.New(reflect.TypeOf(oldObj).Elem())
	oldVal := reflect.ValueOf(oldObj).Elem()
	newVal := newObj.Elem()
	for i := 0; i < oldVal.NumField(); i++ {
		newValField := newVal.Field(i)
		if newValField.CanSet() {
			newValField.Set(oldVal.Field(i))
		}
	}

	return newObj.Interface()
}
