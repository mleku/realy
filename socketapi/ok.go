package socketapi

import (
	"realy.lol/envelopes/eid"
	"realy.lol/envelopes/okenvelope"
	"realy.lol/reason"
)

type OK func(a *A, env eid.Ider, format string, params ...any) (err error)

type OKs struct {
	AuthRequired OK
	PoW          OK
	Duplicate    OK
	Blocked      OK
	RateLimited  OK
	Invalid      OK
	Error        OK
	Unsupported  OK
	Restricted   OK
}

var Ok = OKs{
	AuthRequired: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.AuthRequired.F(format, params...)).Write(a.Listener)
	},
	PoW: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.PoW.F(format, params...)).Write(a.Listener)
	},
	Duplicate: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.Duplicate.F(format, params...)).Write(a.Listener)
	},
	Blocked: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.Blocked.F(format, params...)).Write(a.Listener)
	},
	RateLimited: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.RateLimited.F(format, params...)).Write(a.Listener)
	},
	Invalid: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.Invalid.F(format, params...)).Write(a.Listener)
	},
	Error: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.Error.F(format, params...)).Write(a.Listener)
	},
	Unsupported: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.Unsupported.F(format, params...)).Write(a.Listener)
	},
	Restricted: func(a *A, env eid.Ider, format string, params ...any) (err error) {
		return okenvelope.NewFrom(env.Id(), false, reason.Restricted.F(format, params...)).Write(a.Listener)
	},
}
