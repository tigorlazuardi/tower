package towerhttp

import (
	"encoding/json"
	"fmt"
	"github.com/tigorlazuardi/tower"
	"net/http"
	"strings"
)

func NewLoggerHook(opts ...RespondHookOption) RespondHook {
	return NewRespondHook(append(defaultLoggerOptions(), opts...)...)
}

func defaultLoggerOptions() []RespondHookOption {
	opts := make([]RespondHookOption, 0, 8)
	opts = append(opts, Option.RespondHook().FilterRequest(func(r *http.Request) bool {
		return isHumanReadable(r.Header.Get("Content-Type"))
	}))
	opts = append(opts, Option.RespondHook().ReadRequestBodyLimit(1024*1024))       // 1mb
	opts = append(opts, Option.RespondHook().ReadRespondBodyStreamLimit(1024*1024)) // 1mb
	opts = append(opts, Option.RespondHook().FilterRespondStream(func(respondContentType string, r *http.Request) bool {
		return isHumanReadable(respondContentType)
	}))
	opts = append(opts, Option.RespondHook().OnRespond(defaultLoggerRespond))
	opts = append(opts, Option.RespondHook().OnRespondError(defaultLoggerRespondError))
	opts = append(opts, Option.RespondHook().OnRespondStream(defaultLoggerRespondStream))
	return opts
}

func defaultLoggerRespond(ctx *RespondHookContext) {
	fields := buildLoggerFields(ctx.baseHook, ctx.ResponseBody.PostEncoded, false)
	message := fmt.Sprintf("%s %s %s", ctx.Request.Method, ctx.Request.URL.String(), ctx.Request.Proto)
	if ctx.Error != nil {
		_ = ctx.Tower.Wrap(ctx.Error).Level(tower.ErrorLevel).Code(ctx.ResponseStatus).Message(message).Caller(ctx.Context.Caller).Context(fields).Log(ctx.Request.Context())
		return
	}
	ctx.Tower.NewEntry(message).Level(tower.InfoLevel).Code(ctx.ResponseStatus).Caller(ctx.Context.Caller).Context(fields).Log(ctx.Request.Context())
}

func defaultLoggerRespondError(ctx *RespondErrorHookContext) {
	fields := buildLoggerFields(ctx.baseHook, ctx.ResponseBody.PostEncoded, false)
	message := fmt.Sprintf("%s %s %s", ctx.Request.Method, ctx.Request.URL.String(), ctx.Request.Proto)
	if ctx.Error != nil {
		_ = ctx.Tower.Wrap(ctx.Error).Code(ctx.ResponseStatus).Message(message).Caller(ctx.Context.Caller).Context(fields).Log(ctx.Request.Context())
		return
	}
	_ = ctx.Tower.Wrap(ctx.ResponseBody.PreEncoded).Code(ctx.ResponseStatus).Caller(ctx.Context.Caller).Context(fields).Log(ctx.Request.Context())
}

func defaultLoggerRespondStream(ctx *RespondStreamHookContext) {
	fields := buildLoggerFields(ctx.baseHook, ctx.ResponseBody.Value.CloneBytes(), ctx.ResponseBody.Value.Truncated())
	message := fmt.Sprintf("%s %s %s", ctx.Request.Method, ctx.Request.URL.String(), ctx.Request.Proto)
	if ctx.Error != nil {
		_ = ctx.Tower.Wrap(ctx.Error).Level(tower.ErrorLevel).Code(ctx.ResponseStatus).Message(message).Caller(ctx.Context.Caller).Context(fields).Log(ctx.Request.Context())
		return
	}
	ctx.Tower.NewEntry(message).Level(tower.InfoLevel).Code(ctx.ResponseStatus).Caller(ctx.Context.Caller).Context(fields).Log(ctx.Request.Context())
}

func buildLoggerFields(hook *baseHook, respBody []byte, truncated bool) tower.F {
	url := hook.Request.Host + hook.Request.URL.String()
	requestFields := tower.F{
		"method": hook.Request.Method,
		"url":    url,
	}
	if len(hook.Request.Header) > 0 {
		requestFields["headers"] = hook.Request.Header
	}

	if hook.RequestBody.Len() > 0 {
		contentType := hook.Request.Header.Get("Content-Type")
		switch {
		case hook.RequestBody.Truncated():
			requestFields["body"] = fmt.Sprintf("%s (truncated)", hook.RequestBody.String())
		case strings.Contains(contentType, "application/json") && isJson(hook.RequestBody.Bytes()):
			requestFields["body"] = json.RawMessage(hook.RequestBody.CloneBytes())
		case contentType == "" && isJsonLite(hook.RequestBody.Bytes()) && isJson(hook.RequestBody.Bytes()):
			requestFields["body"] = json.RawMessage(hook.RequestBody.CloneBytes())
		default:
			requestFields["body"] = hook.RequestBody.String()
		}
	}

	responseFields := tower.F{
		"status": hook.ResponseStatus,
	}
	if len(hook.ResponseHeader) > 0 {
		responseFields["headers"] = hook.ResponseHeader
	}
	if len(respBody) > 0 {
		contentType := hook.ResponseHeader.Get("Content-Type")
		switch {
		case truncated:
			responseFields["body"] = fmt.Sprintf("%s (truncated)", hook.RequestBody.String())
		case strings.Contains(contentType, "application/json") && isJson(respBody):
			responseFields["body"] = json.RawMessage(respBody)
		case contentType == "" && isJsonLite(respBody) && isJson(respBody):
			responseFields["body"] = json.RawMessage(respBody)
		default:
			responseFields["body"] = string(respBody)
		}
	}

	return tower.F{
		"request":  requestFields,
		"response": responseFields,
	}
}

func isJsonLite(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	return (b[0] == '{' || b[0] == '[') && (b[len(b)-1] == '}' || b[len(b)-1] == ']')
}
