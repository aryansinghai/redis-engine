package core

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"time"
)

var RESP_NIL = []byte("$-1\r\n")
var RESP_NIL2 = []byte(":-2\r\n")
var RESP_NIL_INTEGER = []byte(":-1\r\n")
var RESP_OK = []byte("+OK\r\n") // #genai: RESP simple strings must start with +

func EvalAndRespond(cmds RedisCmds, c io.ReadWriter) error {

	var response []byte
	buf := bytes.NewBuffer(response)

	for _, cmd := range cmds {
		switch cmd.Cmd {
		case "PING":
			buf.Write(evalPing(cmd.Args, c))
		case "GET":
			buf.Write(evalGet(cmd.Args, c))
		case "SET":
			buf.Write(evalSet(cmd.Args, c))
		case "TTL":
			buf.Write(evalTTL(cmd.Args, c))
		case "DEL":
			buf.Write(evalDel(cmd.Args, c))
		case "EXPIRE":
			buf.Write(evalExpire(cmd.Args, c))
		case "BGREWRITEAOF":
			buf.Write(evalBgRewriteAOF(cmd.Args))
		default:
			buf.Write(evalPing(cmd.Args, c))
		}
	}
	c.Write(buf.Bytes())
	return nil
}

func evalBgRewriteAOF(args []string) []byte {
	DumpAllAOF()
	return RESP_OK
}

func evalPing(args []string, c io.ReadWriter) []byte {
	var b []byte

	if len(args) >= 2 {
		return Encode(errors.New("wrong number of arguments for 'ping' command"), false)
	}

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b
}

func evalSet(args []string, c io.ReadWriter) []byte {
	if len(args) <= 1 {
		return Encode(errors.New("wrong number of arguments for 'set' command"), false)
	}

	var key, value string
	var expTime int64 = -1
	var err error
	var expTimeMs int64 = -1

	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			if i+1 >= len(args) {
				return Encode(errors.New("wrong number of arguments for 'set' command"), false)
			}
			expTime, err = strconv.ParseInt(args[i+1], 10, 64)
			if err != nil {
				return Encode(errors.New("invalid expiration time for 'set' command"), false)
			}
			expTimeMs = expTime * 1000
			i++
		default:
			return Encode(errors.New("invalid argument for 'set' command"), false)
		}
	}

	Put(key, NewObject(value, expTimeMs))
	return RESP_OK
}

func evalGet(args []string, c io.ReadWriter) []byte {
	if len(args) != 1 {
		return Encode(errors.New("wrong number of arguments for 'get' command"), false)
	}

	key := args[0]
	obj := Get(key)
	if obj == nil {
		return RESP_NIL
	}
	return Encode(obj.Value, false)
}

func evalTTL(args []string, c io.ReadWriter) []byte {
	if len(args) != 1 {
		return Encode(errors.New("wrong number of arguments for 'ttl' command"), false)
	}

	key := args[0]
	obj := Get(key)

	if obj == nil {
		return RESP_NIL2
	}

	if obj.ExpiresAt == -1 {
		return RESP_NIL_INTEGER
	}

	ttl := obj.ExpiresAt - time.Now().UnixMilli()
	if ttl <= 0 {
		return RESP_NIL2
	}
	return Encode(int64(ttl/1000), false)
}

func evalDel(args []string, c io.ReadWriter) []byte {
	var countDeleted int = 0
	for _, key := range args {
		if Delete(key) {
			countDeleted++
		}
	}
	return Encode(countDeleted, false)
}

func evalExpire(args []string, c io.ReadWriter) []byte {
	if len(args) < 2 {
		return Encode(errors.New("wrong number of arguments for 'expire' command"), false)
	}

	var key string = args[0]
	expDurationSecs, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return Encode(errors.New("invalid expiration time for 'expire' command"), false)
	}
	obj := Get(key)
	if obj == nil {
		return Encode(0, false)
	}
	expDurationMs := time.Now().UnixMilli() + expDurationSecs*1000
	obj.ExpiresAt = expDurationMs
	return Encode(1, false)
}
