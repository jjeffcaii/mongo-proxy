package middware

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"

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

type SecurityManager interface {
	AsMiddware() pxmgo.Middleware
	Ok() (db *string, ok bool)
}

type simpleVerifier struct {
	step           saslstep
	checker        *scram.Server
	conversationId int32
	getCredential  fnCredential
	db             *string
}

func (p *simpleVerifier) AsMiddware() pxmgo.Middleware {
	return p.validate
}

func (p *simpleVerifier) Ok() (*string, bool) {
	if p.step == success {
		return p.db, true
	}
	return nil, false
}

func NewSecurityManager(fn fnCredential) SecurityManager {
	return &simpleVerifier{
		step:           waitStart,
		checker:        scram.NewServer(sha1.New, gen),
		conversationId: 0,
		getCredential:  fn,
	}
}

func (p *simpleVerifier) extractDB(req *protocol.OpQuery) *string {
	i := strings.Index(req.FullCollectionName, ".")
	db := req.FullCollectionName[:i]
	return &db
}

func (p *simpleVerifier) validate(req protocol.Message, res pxmgo.OnRes, next pxmgo.OnNext) {
	defer func() {
		if err := recover(); err != nil {
			res(mkFailedReply())
			next(pxmgo.END)
		}
	}()

	var q *protocol.OpQuery
	if v, ok := req.(*protocol.OpQuery); ok {
		q = v
	} else if p.step != success {
		panic(errors.New("invalid auth"))
	} else {
		next(nil)
		return
	}
	database := p.extractDB(q)
	w := protocol.ToMap(q.Query)
	switch p.step {
	case waitStart:
		if w["saslStart"] != nil {
			p.saslStart(database, q, res, next)
		} else {
			panic(errors.New("need sasl start"))
		}
		break
	case waitContinue:
		if w["saslContinue"] != nil {
			p.saslContinue(database, q, res, next)
		} else {
			panic(errors.New("need sasl continue"))
		}
		break
	case waitContinue2:
		if w["saslContinue"] != nil {
			p.saslContinue2(database, q, res, next)
		} else {
			panic(errors.New("need sasl continue"))
		}
		break
	case success:
		// 检查要访问的DB是否授权
		if *database == *(p.db) {
			next(nil)
		} else {
			panic(errors.New(fmt.Sprintf("cannot access database %s", *database)))
		}
		break
	case failed:
		panic(errors.New("invalid auth"))
		break
	default:
		panic(errors.New("should never"))
		break
	}
}

func mkFailedReply() protocol.Message {
	body := protocol.NewDocument().
		Set("ok", int64(0)).
		Set("errmsg", "authentication failed").
		Set("code", int32(18)).
		Set("codeName", "AuthenticationFailed").
		Build()
	ret := &protocol.OpReply{
		Header: &protocol.Header{
			OpCode: protocol.OpCodeReply,
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{body},
	}
	return ret
}

func (p *simpleVerifier) saslStart(db *string, req *protocol.OpQuery, res pxmgo.OnRes, next pxmgo.OnNext) {
	defer func() {
		if err := recover(); err != nil {
			res(mkFailedReply())
			p.step = failed
		} else {
			p.step = waitContinue
		}
		next(pxmgo.END)
	}()

	m := protocol.ToMap(req.Query)
	if v, ok := m["mechanism"].(bson.String); !ok || v != bson.String(supportedMechanism) {
		panic(fmt.Errorf("invalid mechanism: expect=%s", supportedMechanism))
	}
	var payload []byte
	if v, ok := (m["payload"]).(bson.Binary); ok {
		payload = []byte(v)
	} else {
		panic(errors.New("invalid payload"))
	}
	if err := p.checker.ParseClientFirst(payload); err != nil {
		panic(err)
	}
	if username, _, err := p.getCredential(db); err != nil {
		panic(err)
	} else if p.checker.UserName() != *username {
		panic(fmt.Errorf("invalid username %s", p.checker.UserName()))
	} else {
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
			Header: &protocol.Header{
				ResponseTo: req.GetHeader().RequestID,
				OpCode:     protocol.OpCodeReply,
			},
			ResponseFlags:  8,
			NumberReturned: 1,
			Documents:      []protocol.Document{doc},
		}
		res(rep)
	}
}

func (p *simpleVerifier) saslContinue(db *string, req *protocol.OpQuery, res pxmgo.OnRes, next pxmgo.OnNext) {
	defer func() {
		if err := recover(); err != nil {
			res(mkFailedReply())
			p.step = failed
		} else {
			p.step = waitContinue2
		}
		next(pxmgo.END)
	}()
	m := protocol.ToMap(req.Query)
	if cid, ok := m["conversationId"].(bson.Int32); !ok || int32(cid) != p.conversationId {
		panic(errors.New("invalid conversationId"))
	}
	var payload []byte
	if b, ok := m["payload"].(bson.Binary); ok {
		payload = b
	} else {
		panic(errors.New("invalid payload"))
	}
	username, password, err := p.getCredential(db)
	if err != nil {
		panic(err)
	}
	pwd := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:mongo:%s", *username, *password))))
	p.checker.SaltPassword([]byte(pwd))
	if err := p.checker.CheckClientFinal(payload); err != nil {
		log.Println("check client final faile:", err)
		panic(err)
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
		Header: &protocol.Header{
			OpCode: protocol.OpCodeReply,
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{doc},
	}
	res(rep)
}

func (p *simpleVerifier) saslContinue2(db *string, req *protocol.OpQuery, res pxmgo.OnRes, next pxmgo.OnNext) {
	defer func() {
		if err := recover(); err != nil {
			res(mkFailedReply())
			p.step = failed
		} else {
			p.step = success
		}
		next(pxmgo.END)
	}()

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
		Header: &protocol.Header{
			OpCode: protocol.OpCodeReply,
		},
		ResponseFlags:  8,
		NumberReturned: 1,
		Documents:      []protocol.Document{doc},
	}
	res(rep)
	p.db = db
}

type stdGenerator struct{}

var gen = &stdGenerator{}

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
