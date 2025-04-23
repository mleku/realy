package socketapi

import (
	"realy.mleku.dev/envelopes/eventenvelope"
	"realy.mleku.dev/envelopes/okenvelope"
	"realy.mleku.dev/reason"
)

func (a *A) Ok(format string, prefix reason.R, env *eventenvelope.Submission, params ...any) (err error) {
	err = okenvelope.NewFrom(env.Id, false, prefix.F(format, params...)).Write(a.Listener)
	return
}

func (a *A) AuthRequired(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.AuthRequired.F(format, params...)).Write(a.Listener)
}

func (a *A) PoW(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.PoW.F(format, params...)).Write(a.Listener)
}

func (a *A) Duplicate(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.Duplicate.F(format, params...)).Write(a.Listener)
}

func (a *A) Blocked(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.Blocked.F(format, params...)).Write(a.Listener)
}

func (a *A) RateLimited(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.RateLimited.F(format, params...)).Write(a.Listener)
}

func (a *A) Invalid(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.Invalid.F(format, params...)).Write(a.Listener)
}

func (a *A) Error(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.Error.F(format, params...)).Write(a.Listener)
}

func (a *A) Unsupported(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.Unsupported.F(format, params...)).Write(a.Listener)
}

func (a *A) Restricted(env *eventenvelope.Submission, format string, params ...any) (err error) {
	return okenvelope.NewFrom(env.Id, false, reason.Restricted.F(format, params...)).Write(a.Listener)
}
