package smtpclient

import (
	"fmt"
	"net/smtp"
)

type Client struct {
	Host   string
	Port   int
	Sender string
}

type Options struct {
	Host   string
	Port   int
	Sender string
}

func NewClient(opts *Options) *Client {
	return &Client{
		Host:   opts.Host,
		Port:   opts.Port,
		Sender: opts.Sender,
	}
}

func (c *Client) Send(to []string, msg string) error {
	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)

	return smtp.SendMail(addr, nil, c.Sender, to, []byte(msg))
}
