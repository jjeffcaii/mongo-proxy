package middware

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/goxmpp/sasl/scram"
	"github.com/jjeffcaii/mongo-proxy"
	"github.com/jjeffcaii/mongo-proxy/protocol"
	"github.com/sbunce/bson"
)

type saslstep uint8

const (
	supportedMechanism = "SCRAM-SHA-1"

	waitStart     saslstep = iota
	waitContinue
	waitContinue2
	success
	failed
)

var (
	failedResponse *protocol.OpReply
	errAuthReq     = errors.New("auth failed: illegal auth request")
	errAuthFailed  = errors.New("auth failed")
)

func init() {
	body := protocol.NewDocument().
		Set("ok", int64(0)).
		Set("errmsg", "authentication failed").
		Set("code", int32(18)).
		Set("codeName", "AuthenticationFailed").
		Build()
	failedResponse = &protocol.OpReply{
		Op: &protocol.Op{
			OpHeader: &protocol.Header{
				OpCode: protocol.OpCodeReply,
			},
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{body},
	}
}

type simpleAuthenticator struct {
	step           saslstep
	checker        *scram.Server
	conversationId int32
	getCredential  func(string) (*pxmgo.Identifier, error)
	db             *string
	notifier       *sync.WaitGroup
}

func (p *simpleAuthenticator) Wait() (db *string, ok bool) {
	if p.step == failed {
		return nil, false
	}
	if p.step == success {
		return p.db, true
	}
	p.notifier.Wait()
	if p.step == success {
		return p.db, true
	}
	return nil, false
}

func (p *simpleAuthenticator) Handle(ctx pxmgo.Context, req protocol.Message) error {
	err := p.auth(ctx, req)
	if err == nil || err == pxmgo.EOF || err == pxmgo.Ignore {
		return err
	}
	log.Println("auth failed:", err)
	p.step = failed
	p.notifier.Done()
	if err := ctx.SendMessage(failedResponse); err != nil {
		log.Println("send AuthFailed response error:", err)
	}
	return err
}

func (p *simpleAuthenticator) auth(ctx pxmgo.Context, req protocol.Message) error {
	q, ok := req.(*protocol.OpQuery)
	if !ok {
		return errAuthReq
	}
	tbl, ok := q.TableName()
	if !ok {
		return errAuthReq
	}
	if p.step == success {
		// 检查要访问的DB是否授权
		if tbl.Database == *(p.db) {
			return nil
		}
		if tbl.Database == "admin" {
			log.Println("[WARN] access admin database")
			return nil
		}
		return &errDenyDB{tbl.Database}
	}
	if p.step == failed {
		return errAuthFailed
	}
	if p.step == waitStart {
		if _, ok := protocol.Load(q.Query, "saslStart"); !ok {
			return errAuthReq
		}
		err := p.saslStart(ctx, tbl.Database, q)
		if err != nil {
			return err
		}
		p.step = waitContinue
		return pxmgo.Ignore
	}
	if p.step == waitContinue {
		if _, ok := protocol.Load(q.Query, "saslContinue"); !ok {
			return errAuthReq
		}
		if err := p.saslContinue(ctx, tbl.Database, q); err != nil {
			return err
		}
		p.step = waitContinue2
		return pxmgo.Ignore
	}
	if p.step == waitContinue2 {
		if _, ok := protocol.Load(q.Query, "saslContinue"); !ok {
			return errAuthReq
		}
		if err := p.saslContinue2(ctx, tbl.Database, q); err != nil {
			return err
		}
		p.step = success
		p.notifier.Done()
		return pxmgo.Ignore
	}
	return errAuthReq
}

func (p *simpleAuthenticator) saslStart(ctx pxmgo.Context, db string, req *protocol.OpQuery) error {
	m := protocol.ToMap(req.Query)
	if v, ok := m["mechanism"].(bson.String); !ok || v != bson.String(supportedMechanism) {
		return fmt.Errorf("invalid mechanism: expect=%s", supportedMechanism)
	}
	var payload []byte
	if v, ok := (m["payload"]).(bson.Binary); ok {
		payload = []byte(v)
	} else {
		return errAuthReq
	}
	if err := p.checker.ParseClientFirst(payload); err != nil {
		return err
	}
	identifier, err := p.getCredential(db)
	if err != nil {
		return err
	}
	if p.checker.UserName() != identifier.Username {
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

func (p *simpleAuthenticator) saslContinue(ctx pxmgo.Context, db string, req *protocol.OpQuery) error {
	m := protocol.ToMap(req.Query)
	if cid, ok := m["conversationId"].(bson.Int32); !ok || int32(cid) != p.conversationId {
		return errAuthReq
	}
	var payload []byte
	if b, ok := m["payload"].(bson.Binary); ok {
		payload = b
	} else {
		return errAuthReq
	}
	identifier, err := p.getCredential(db)
	if err != nil {
		return err
	}
	pwd := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:mongo:%s", identifier.Username, identifier.Password))))
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

func (p *simpleAuthenticator) saslContinue2(ctx pxmgo.Context, db string, req *protocol.OpQuery) error {
	v, ok := protocol.Load(req.Query, "conversationId")
	if !ok {
		return errAuthReq
	}
	if vv, ok := v.(bson.Int32); !ok || p.conversationId != int32(vv) {
		return errAuthReq
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
	p.db = &db
	ctx.SendMessage(rep)
	return nil
}

func NewAuthenticator(fn func(db string) (*pxmgo.Identifier, error)) pxmgo.Authenticator {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return &simpleAuthenticator{
		step:           waitStart,
		checker:        scram.NewServer(sha1.New, gen),
		conversationId: 0,
		getCredential:  fn,
		notifier:       wg,
	}
}

var gen = &stdGenerator{}

type stdGenerator struct{}

func (g *stdGenerator) GetNonce(ln int) []byte {
	if ln == 21 {
		return []byte("fyko+d2lbbFgONRv9qkxdawL") // Client's nonce
	}
	return []byte("3rfcNHYJY1ZVvWVs7j") // Server's nonce
}

func (g *stdGenerator) GetSalt(ln int) []byte {
	b, err := base64.StdEncoding.DecodeString("QSXCR+Q6sek8bf92")
	if err != nil {
		panic(err)
	}
	return b
}

func (g *stdGenerator) GetIterations() int {
	return 4096
}

type errDenyDB struct {
	db string
}

func (p *errDenyDB) Error() string {
	return fmt.Sprintf("access deny: db=%s", p.db)
}
