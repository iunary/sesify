package sender

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/iunary/sesify/internal/compaign"
)

type SES struct {
	Region string
	SESID  string
	SESKEY string
}

func NewSES(region, sesid, seskey string) *SES {
	return &SES{
		Region: region,
		SESID:  sesid,
		SESKEY: seskey,
	}
}

func (s *SES) Send(ctx context.Context, comp *compaign.Compaign) compaign.Result {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(s.Region),
		Credentials: credentials.NewStaticCredentials(s.SESID, s.SESKEY, ""),
	})

	if err != nil {
		return compaign.Result{
			Email:     comp.Recepient.Email,
			Error:     err,
			Delivered: false,
		}
	}
	// compile the compaign template
	if err := comp.Render(); err != nil {
		return compaign.Result{
			Email:     comp.Recepient.Email,
			Error:     err,
			Delivered: false,
		}
	}
	svc := ses.New(sess)
	input := &ses.SendEmailInput{
		// Set destination emails
		Destination: &ses.Destination{
			ToAddresses: []*string{&comp.Recepient.Email},
		},

		// Set email message and subject
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(comp.Message),
				},
			},

			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(comp.Subject),
			},
		},

		// send from email
		Source: aws.String(comp.From),
	}

	_, err = svc.SendEmail(input)
	if err != nil {
		return compaign.Result{
			Email:     comp.Recepient.Email,
			Error:     err,
			Delivered: false,
		}
	}
	return compaign.Result{
		Email:     comp.Recepient.Email,
		Error:     nil,
		Delivered: true,
	}
}
