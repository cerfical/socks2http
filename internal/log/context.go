package log

import "github.com/rs/zerolog"

type Context struct {
	ctx zerolog.Context
}

func (c *Context) Str(key, val string) *Context {
	return &Context{c.ctx.Str(key, val)}
}

func (c *Context) Level(lvl Level) *Context {
	return &Context{c.ctx.Logger().Level(zerolog.Level(lvl)).With()}
}

func (c *Context) Logger() *Logger {
	return &Logger{c.ctx.Logger()}
}
