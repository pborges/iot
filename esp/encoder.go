package esp

import (
	"errors"
	"fmt"
	"strings"
)

type Command struct {
	Name string
	Args map[string]string
}

func Encode(c Command) string {
	s := make([]string, 0, len(c.Args)+1)
	s = append(s, sanitize(c.Name))
	for k, v := range c.Args {
		s = append(s, encodeKey(k, v))
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

func Decode(str string) (Command, error) {
	var state tokenizerState
	var cmd Command
	cmd.Args = make(map[string]string)
	var key string
	var value string
	var inQuote bool
	for _, c := range str {
		switch state {
		case tokDecodeCommand:
			if c != ' ' {
				cmd.Name = cmd.Name + string(c)
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
					return cmd, errors.New("unexpected space in key")
				}
				key += string(c)
			case ':':
				if inQuote {
					return cmd, errors.New("unclosed quote")
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
					cmd.Args[key] = value
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
	cmd.Args[key] = value
	return cmd, nil
}
