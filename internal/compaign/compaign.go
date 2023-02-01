package compaign

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

type Recipient struct {
	UUID      string
	Firstname string
	Lastname  string
	Email     string
}

type Compaign struct {
	Subject    string
	From       string
	Message    string
	Recepient  *Recipient
	Attachment string
	Delay      time.Duration
	tpl        *template.Template
	subjectTpl *template.Template
}

func (c *Compaign) Render() error {
	if err := c.compileTemplate(); err != nil {
		return err
	}
	out := bytes.Buffer{}
	if c.subjectTpl != nil {
		err := c.subjectTpl.ExecuteTemplate(&out, contentTpl, c.Recepient)
		if err != nil {
			return err
		}
		c.Subject = out.String()
		out.Reset()
	}

	if err := c.tpl.ExecuteTemplate(&out, contentTpl, c.Recepient); err != nil {
		return err
	}
	c.Message = out.String()
	return nil
}

func (c *Compaign) compileTemplate() error {
	body := c.Message
	baseTPL, err := template.New(baseTpl).Parse(body)
	if err != nil {
		return fmt.Errorf("error compiling base template: %v", err)
	}
	body = c.Message

	msgTpl, err := template.New(contentTpl).Parse(body)
	if err != nil {
		return fmt.Errorf("error compiling message: %v", err)
	}

	out, err := baseTPL.AddParseTree(contentTpl, msgTpl.Tree)
	if err != nil {
		return fmt.Errorf("error inserting child template: %v", err)
	}
	c.tpl = out

	// If the subject line has a template string, compile it.
	if strings.Contains(c.Subject, "{{") {
		subj := c.Subject
		subjTpl, err := template.New(contentTpl).Parse(subj)
		if err != nil {
			return fmt.Errorf("error compiling subject: %v", err)
		}
		c.subjectTpl = subjTpl
	}
	return nil
}

type Result struct {
	Email     string
	Delivered bool
	Error     error
}

type Stats struct {
	Delivered int
	Failed    int
}
