package espiot

import (
	"errors"
	"fmt"
	"strings"
)

type Command struct {
	Name string
	Args map[string]string
}

func Encode(p Packet) string {
	s := make([]string, 0, len(p.Args)+1)
	s = append(s, sanitize(p.Command))
	if p.Args != nil {
		for k, v := range p.Args {
			s = append(s, encodeKey(k, v))
		}
	}
	return fmt.Sprint(strings.Join(s, " "))
}

func encodeKey(key, value string) string {
	return fmt.Sprintf("%s:%s", sanitize(key), sanitize(value))
}

func sanitize(s string) string {
	if strings.Contains(s, " ") {
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}

type tokenizerState int

var tokDecodeCommand tokenizerState = 0
var tokDecodeKey tokenizerState = 1
var tokDecodeValue tokenizerState = 2

func Decode(str string) (Packet, error) {
	var state tokenizerState
	var p Packet
	p.Args = make(map[string]string)
	var key string
	var value string
	var inQuote bool
	for _, c := range str {
		switch state {
		case tokDecodeCommand:
			if c != ' ' {
				p.Command = p.Command + string(c)
			} else {
				state++
			}
		case tokDecodeKey:
			switch c {
			case '"':
				if inQuote {
					inQuote = false
				} else {
					inQuote = true
				}
			case ' ':
				if !inQuote {
					return p, errors.New("unexpected space in key")
				}
				key += string(c)
			case ':':
				if inQuote {
					return p, errors.New("unclosed quote")
				}
				inQuote = false
				state = tokDecodeValue
			default:
				key += string(c)
			}
		case tokDecodeValue:
			switch c {
			case '"':
				if inQuote {
					inQuote = false
				} else {
					inQuote = true
				}
			case ' ':
				if !inQuote {
					p.Args[key] = value
					key, value = "", ""
					inQuote = false
					state = tokDecodeKey
				} else {
					value += string(c)
				}
			default:
				value += string(c)
			}
		}
	}
	p.Args[key] = value
	return p, nil
}
