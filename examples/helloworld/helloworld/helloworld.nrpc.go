// This code was autogenerated from helloworld.proto, do not edit.
package helloworld

import (
	"context"
	"log"
	"time"

	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-rpc/nrpc"
)

// GreeterServer is the interface that providers of the service
// Greeter should implement.
type GreeterServer interface {
	SayHello(ctx context.Context, req HelloRequest) (resp HelloReply, err error)
}

// GreeterHandler provides a NATS subscription handler that can serve a
// subscription using a given GreeterServer implementation.
type GreeterHandler struct {
	ctx    context.Context
	nc     nrpc.NatsConn
	server GreeterServer
}

func NewGreeterHandler(ctx context.Context, nc nrpc.NatsConn, s GreeterServer) *GreeterHandler {
	return &GreeterHandler{
		ctx:    ctx,
		nc:     nc,
		server: s,
	}
}

func (h *GreeterHandler) Subject() string {
	return "helloworld.Greeter.>"
}

func (h *GreeterHandler) Handler(msg *nats.Msg) {
	request := nrpc.NewRequest(msg.Subject, msg.Reply)
	// extract method name & encoding from subject
	_, _, name, tail, err := nrpc.ParseSubject(
		"helloworld", 0, "Greeter", 0, msg.Subject)
	if err != nil {
		log.Printf("GreeterHanlder: Greeter subject parsing failed: %v", err)
		return
	}

	request.MethodName = name
	request.SubjectTail = tail

	ctx := context.WithValue(h.ctx, nrpc.RequestContextKey, request)

	// call handler and form response
	var resp proto.Message
	var replyError *nrpc.Error
	switch name {
	case "SayHello":
		_, request.Encoding, err = nrpc.ParseSubjectTail(0, request.SubjectTail)
		if err != nil {
			log.Printf("SayHelloHanlder: SayHello subject parsing failed: %v", err)
			break
		}
		var req HelloRequest
		if err := nrpc.Unmarshal(request.Encoding, msg.Data, &req); err != nil {
			log.Printf("SayHelloHandler: SayHello request unmarshal failed: %v", err)
			replyError = &nrpc.Error{
				Type: nrpc.Error_CLIENT,
				Message: "bad request received: " + err.Error(),
			}
		} else {
			request.Handler = func()(proto.Message, error){
				innerResp, err := h.server.SayHello(ctx, req)
				if err != nil {
					return nil, err
				}
				return &innerResp, err
			}
			resp, replyError = request.Run()
			if replyError != nil {
				log.Printf("SayHelloHandler: SayHello handler failed: %s", replyError.Error())
			}
		}
	default:
		log.Printf("GreeterHandler: unknown name %q", name)
		replyError = &nrpc.Error{
			Type: nrpc.Error_CLIENT,
			Message: "unknown name: " + name,
		}
	}


	if !request.NoReply {
		// encode and send response
		err = nrpc.Publish(resp, replyError, h.nc, msg.Reply, request.Encoding) // error is logged
	} else {
		err = nil
	}
	if err != nil {
		log.Println("GreeterHandler: Greeter handler failed to publish the response: %s", err)
	}
}

type GreeterClient struct {
	nc      nrpc.NatsConn
	PkgSubject string
	Subject string
	Encoding string
	Timeout time.Duration
}

func NewGreeterClient(nc nrpc.NatsConn) *GreeterClient {
	return &GreeterClient{
		nc:      nc,
		PkgSubject: "helloworld",
		Subject: "Greeter",
		Encoding: "protobuf",
		Timeout: 5 * time.Second,
	}
}

func (c *GreeterClient) SayHello(req HelloRequest) (resp HelloReply, err error) {

	subject := c.PkgSubject + "." + c.Subject + "." + "SayHello"

	// call
	err = nrpc.Call(&req, &resp, c.nc, subject, c.Encoding, c.Timeout)
	if err != nil {
		return // already logged
	}

	return
}

type Client struct {
	nc      nrpc.NatsConn
	defaultEncoding string
	defaultTimeout time.Duration
	pkgSubject string
	Greeter *GreeterClient
}

func NewClient(nc nrpc.NatsConn) *Client {
	c := Client{
		nc: nc,
		defaultEncoding: "protobuf",
		defaultTimeout: 5*time.Second,
		pkgSubject: "helloworld",
	}
	c.Greeter = NewGreeterClient(nc)
	return &c
}

func (c *Client) SetEncoding(encoding string) {
	c.defaultEncoding = encoding
	if c.Greeter != nil {
		c.Greeter.Encoding = encoding
	}
}

func (c *Client) SetTimeout(t time.Duration) {
	c.defaultTimeout = t
	if c.Greeter != nil {
		c.Greeter.Timeout = t
	}
}