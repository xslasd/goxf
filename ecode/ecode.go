package ecode

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

type ECodes interface {
	// Error sometimes Error return Status in string form
	// NOTE: don't use Error in monitor report even it also work for now
	Error() string
	// Code Status get error code.
	Code() int
	// Message get code message.
	Message() string
	//Values get formatted message parameter,it may be nil.
	Values() []string
	// Details Detail get error detail,it may be nil.
	Details() []any

	SetMessage(msg string) ECodes
	SetMessagef(f ...string) ECodes
	WithDetails(msg any) ECodes
}

var (
	_messages atomic.Value           // NOTE: stored map[string]map[int]string
	_codes    = make(map[int]string) // register codes.
	mu        sync.Mutex
)

func Register(cm map[int]string) {
	mu.Lock()
	defer mu.Unlock()
	var codes = make(map[int]string)
	rcm, ok := _messages.Load().(map[int]string)
	if ok {
		if cm != nil {
			for k, v := range rcm {
				codes[k] = v
			}
		}
	}

	if cm != nil {
		for k, v := range cm {
			codes[k] = v
		}
	}
	_messages.Store(codes)
}

func GetRegisteredCode() map[int]string {
	var codes = make(map[int]string)
	cm, has := _messages.Load().(map[int]string)
	if has {
		for k, v := range cm {
			codes[k] = v
		}
		return cm
	}
	return nil
}

func GetAllCodes() map[int]string {
	var codes = make(map[int]string)
	cm, has := _messages.Load().(map[int]string)
	for k, v := range _codes {
		if has {
			if mv, ok := cm[k]; ok {
				codes[k] = mv
			} else {
				codes[k] = v
			}
		} else {
			codes[k] = v
		}
	}
	return codes
}

func New(e int, msg string) ECodes {
	if e <= 0 {
		panic("business ecode must greater than zero")
	}
	return add(e, msg)
}

func add(e int, msg string) ECodes {
	if _, ok := _codes[e]; ok {
		panic(fmt.Sprintf("ecode: %d already exist", e))
	}
	_codes[e] = msg
	return &ECode{id: e, msg: msg}
}

type ECode struct {
	id     int
	msg    string
	f      []string
	detail []any
}

func (c *ECode) Error() string {
	return strconv.Itoa(c.id)
}
func (c *ECode) Code() int { return c.id }

func (c *ECode) Values() []string {
	return c.f
}
func (c *ECode) Message() string {
	if c.msg != "" {
		if c.f != nil {
			var anySlice []any
			for _, s := range c.f {
				anySlice = append(anySlice, s)
			}
			return fmt.Sprintf(c.msg, anySlice...)
		}
		return c.msg
	}
	cm, ok := _messages.Load().(map[int]string)
	if ok {
		if msg, ok := cm[c.id]; ok {
			if c.f != nil {
				var anySlice []any
				for _, s := range c.f {
					anySlice = append(anySlice, s)
				}
				return fmt.Sprintf(msg, anySlice...)
			}
			return msg
		}
	}
	return c.msg
}
func (c *ECode) Details() []any {
	return c.detail
}

func (c *ECode) WithDetails(d any) ECodes {
	return &ECode{id: c.id, msg: c.msg, f: c.f, detail: append(c.detail, d)}
}

func (c *ECode) SetMessage(msg string) ECodes {
	return &ECode{id: c.id, msg: msg, f: c.f, detail: c.detail}
}

func (c *ECode) SetMessagef(f ...string) ECodes {
	return &ECode{id: c.id, msg: c.msg, f: f, detail: c.detail}
}

func Cause(e error) ECodes {
	if e == nil {
		return OK
	}
	ec, ok := errors.Cause(e).(ECodes)
	if ok {
		return ec
	}
	return ServerErr.WithDetails(e.Error())
}

func As(e error) ECodes {
	if e == nil {
		return OK
	}
	var c *ECode
	if errors.As(e, &c) {
		return c
	}
	return ServerErr.WithDetails(e.Error())
}
