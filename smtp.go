package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/emersion/go-smtp"
	"io"
	"log"
	"time"
)

//arguments
var args struct {
	Host          string `help:"SMTP Host"`
	Port          int    `default:"10025" help:"SMTP port"`
	NoAuth        bool   `help:"disable SMTP authentication"`
	PlainAuth     bool   `help:"enable SMTP PLAIN authentication"`
	AnonAuth      bool   `help:"enable SMTP anonymous authentication"`
	User          string `help:"SMTP username"`
	Password      string `help:"SMTP password"`
	Region        string `arg:"required" help:"AWS region (a.g. eu-west-1)"`
	SourceArn     string `help:"AWS Source ARN"`
	FromArn       string `help:"AWS From ARN"`
	ReturnPathArn string `help:"AWS Return Path ARN"`
	AccessKey     string `help:"AWS Access Key"`
	SecretKey     string `help:"AWS Secret Key"`
}

// The Backend implements SMTP server methods.
type SMTP_Backend struct{}

type SMTP_Session struct {
	from string
	to   string
}

// Login handles a login command with username and password.
func (bkd *SMTP_Backend) Login(state *smtp.ConnectionState, username, password string) (smtp.Session, error) {
	if username != args.User || password != args.Password {
		return nil, errors.New("Invalid username or password")
	}
	return &SMTP_Session{}, nil
}

// AnonymousLogin requires clients to authenticate using SMTP AUTH before sending emails
func (bkd *SMTP_Backend) AnonymousLogin(state *smtp.ConnectionState) (smtp.Session, error) {
	if args.AnonAuth {
		return &SMTP_Session{}, nil
	} else {
		return nil, smtp.ErrAuthRequired
	}
}

func (s *SMTP_Session) Mail(from string, opts smtp.MailOptions) error {
	s.from = from
	return nil
}

func (s *SMTP_Session) Rcpt(to string) error {
	s.to = to
	return nil
}

func (s *SMTP_Session) Data(stream io.Reader) error {
	log.Printf("message from: %s to: %s", s.from, s.to)

	aws_conf := &aws.Config{
		Region: aws.String(args.Region),
	}

	if args.AccessKey != "" {
		aws_conf.Credentials = credentials.NewStaticCredentials(args.AccessKey, args.SecretKey, "")
	}

	aws_sess, err := session.NewSession(aws_conf)

	if err != nil {
		log.Fatal(err)
	}

	ses_ins := ses.New(aws_sess)

	email := &ses.SendRawEmailInput{
		Destinations: []*string{aws.String(s.to)},
		Source:       aws.String(s.from),
		RawMessage:   &ses.RawMessage{Data: StreamToByte(stream)},
	}

	if args.FromArn != "" {
		email.FromArn = aws.String(args.FromArn)
	}

	if args.SourceArn != "" {
		email.SourceArn = aws.String(args.SourceArn)
	}

	if args.ReturnPathArn != "" {
		email.ReturnPathArn = aws.String(args.ReturnPathArn)
	}

	_, err = ses_ins.SendRawEmail(email)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				log.Fatalf("AWS:"+ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				log.Fatalf("AWS:"+ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				log.Fatalf("AWS:"+ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				log.Fatal("AWS:" + aerr.Error())
			}
		} else {
			log.Fatal("AWS: " + err.Error())
		}
	}

	return nil
}

func (s *SMTP_Session) Reset() {}

func (s *SMTP_Session) Logout() error {
	return nil
}

func main() {

	arg.MustParse(&args)

	backend := &SMTP_Backend{}

	smtp := smtp.NewServer(backend)

	smtp.Addr = fmt.Sprintf("%s:%d", args.Host, args.Port)
	smtp.Domain = "localhost"
	smtp.ReadTimeout = 100 * time.Second
	smtp.WriteTimeout = 100 * time.Second
	smtp.MaxMessageBytes = 1024 * 1024
	smtp.MaxRecipients = 50

	if args.PlainAuth {
		smtp.AllowInsecureAuth = true
	}

	if args.NoAuth {
		smtp.AuthDisabled = true
	}

	log.Println("Starting server at", smtp.Addr)

	if err := smtp.ListenAndServe(); err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}
