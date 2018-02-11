package pxmgo

import (
	"crypto/md5"
	"crypto/sha1"
	"errors"
	"fmt"
	"log"

	"github.com/goxmpp/sasl/scram"
	"github.com/jjeffcaii/mongo-proxy"
	"github.com/jjeffcaii/mongo-proxy/protocol"
	"github.com/sbunce/bson"
)

type saslstep uint8

const (
	waitStart     saslstep = iota
	waitContinue
	waitContinue2
	success
	failed
)

const (
	supportedMechanism = "SCRAM-SHA-1"
)

type fnCredential = func(database *string) (username *string, password *string, err error)

type simpleVerifier struct {
	step           saslstep
	checker        *scram.Server
	conversationId int32
	getCredential  fnCredential
	db             *string
}

func (p *simpleVerifier) Handle(ctx pxmgo.Context, req protocol.Message) error {
	q, ok := req.(*protocol.OpQuery)
	if !ok {
		return errors.New("invalid auth")
	}
	database := q.GetDatabase()
	w := protocol.ToMap(q.Query)
	switch p.step {
	default:
		return fmt.Errorf("invalid security check step: %d", p.step)
	case waitStart:
		if _, ok := w["saslStart"]; !ok {
			p.sendAuthFailed(ctx, nil)
			return errors.New("need sasl start")
		}
		err := p.saslStart(ctx, database, q)
		if err != nil {
			p.step = failed
			p.sendAuthFailed(ctx, err)
			return err
		}
		p.step = waitContinue
		return pxmgo.Ignore
	case waitContinue:
		if _, ok := w["saslContinue"]; !ok {
			p.sendAuthFailed(ctx, nil)
			return errors.New("need sasl continue")
		}
		err := p.saslContinue(ctx, database, q)
		if err != nil {
			p.step = failed
			p.sendAuthFailed(ctx, err)
			return err
		}
		p.step = waitContinue2
		return pxmgo.Ignore
	case waitContinue2:
		if _, ok := w["saslContinue"]; !ok {
			p.sendAuthFailed(ctx, nil)
			return errors.New("need sasl continue")
		}
		err := p.saslContinue2(ctx, database, q)
		if err != nil {
			p.step = failed
			p.sendAuthFailed(ctx, err)
			return err
		}
		p.step = success
		return pxmgo.Ignore
	case success:
		// 检查要访问的DB是否授权
		if *database == *(p.db) {
			return nil
		}
		p.sendAuthFailed(ctx, fmt.Errorf("cannot access database %s", *database))
		return fmt.Errorf("cannot access database %s", *database)
	case failed:
		return errors.New("invalid auth")
	}
}

func (p *simpleVerifier) Ok() (*string, bool) {
	if p.step == success {
		return p.db, true
	}
	return nil, false
}

func (p *simpleVerifier) sendAuthFailed(ctx pxmgo.Context, err error) error {
	var emsg = func(e error) string {
		if e == nil {
			return "authentication failed"
		}
		return fmt.Sprintf("authentication failed: %s", e.Error())
	}(err)
	body := protocol.NewDocument().
		Set("ok", int64(0)).
		Set("errmsg", emsg).
		Set("code", int32(18)).
		Set("codeName", "AuthenticationFailed").
		Build()
	return ctx.SendMessage(&protocol.OpReply{
		Op: &protocol.Op{
			OpHeader: &protocol.Header{
				OpCode: protocol.OpCodeReply,
			},
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{body},
	})
}

func (p *simpleVerifier) saslStart(ctx pxmgo.Context, db *string, req *protocol.OpQuery) error {
	m := protocol.ToMap(req.Query)
	if v, ok := m["mechanism"].(bson.String); !ok || v != bson.String(supportedMechanism) {
		return fmt.Errorf("invalid mechanism: expect=%s", supportedMechanism)
	}
	var payload []byte
	if v, ok := (m["payload"]).(bson.Binary); ok {
		payload = []byte(v)
	} else {
		return errors.New("invalid payload")
	}
	if err := p.checker.ParseClientFirst(payload); err != nil {
		return err
	}
	username, _, err := p.getCredential(db)
	if err != nil {
		return err
	}
	if p.checker.UserName() != *username {
		return fmt.Errorf("invalid username %s", p.checker.UserName())
	}
	var s1 = p.checker.First()
	// send response
	p.conversationId++
	doc := protocol.NewDocument().
		Set("conversationId", p.conversationId).
		Set("done", false).
		Set("payload", s1).
		Set("ok", float64(1)).
		Build()
	rep := &protocol.OpReply{
		Op: &protocol.Op{
			OpHeader: &protocol.Header{
				ResponseTo: req.Header().RequestID,
				OpCode:     protocol.OpCodeReply,
			},
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{doc},
	}
	return ctx.SendMessage(rep)
}

func (p *simpleVerifier) saslContinue(ctx pxmgo.Context, db *string, req *protocol.OpQuery) error {
	m := protocol.ToMap(req.Query)
	if cid, ok := m["conversationId"].(bson.Int32); !ok || int32(cid) != p.conversationId {
		return errors.New("invalid conversationId")
	}
	var payload []byte
	if b, ok := m["payload"].(bson.Binary); ok {
		payload = b
	} else {
		return errors.New("invalid payload")
	}
	username, password, err := p.getCredential(db)
	if err != nil {
		return err
	}
	pwd := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:mongo:%s", *username, *password))))
	p.checker.SaltPassword([]byte(pwd))
	if err := p.checker.CheckClientFinal(payload); err != nil {
		log.Println("check client final faile:", err)
		return err
	}
	s2 := p.checker.Final()
	// send server final
	doc := protocol.NewDocument().
		Set("conversationId", p.conversationId).
		Set("done", false).
		Set("payload", s2).
		Set("ok", float64(1)).
		Build()
	rep := &protocol.OpReply{
		Op: &protocol.Op{
			OpHeader: &protocol.Header{
				OpCode: protocol.OpCodeReply,
			},
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{doc},
	}
	return ctx.SendMessage(rep)
}

func (p *simpleVerifier) saslContinue2(ctx pxmgo.Context, db *string, req *protocol.OpQuery) error {
	m := protocol.ToMap(req.Query)
	if v, ok := m["conversationId"].(bson.Int32); !ok || int32(v) != p.conversationId {
		panic(errors.New("invalid conversationId"))
	}
	doc := protocol.NewDocument().
		Set("conversationId", p.conversationId).
		Set("done", true).
		Set("payload", make([]byte, 0)).
		Set("ok", float64(1)).
		Build()
	rep := &protocol.OpReply{
		Op: &protocol.Op{
			OpHeader: &protocol.Header{
				OpCode: protocol.OpCodeReply,
			},
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{doc},
	}
	p.db = db
	ctx.SendMessage(rep)
	return nil
}

func NewSecurityManager(fn fnCredential) pxmgo.Middleware {
	return &simpleVerifier{
		step:           waitStart,
		checker:        scram.NewServer(sha1.New, gen),
		conversationId: 0,
		getCredential:  fn,
	}
}
