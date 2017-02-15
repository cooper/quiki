// Copyright (c) 2017, Mitchell Cooper
package wikiclient

// a Message represents a wikiserver message. the Message struct is used both
// for outgoing messages (client requests) and incoming messages (replies from
// the wikiserver).

import (
	"encoding/json"
	"errors"
	"fmt"
)

type messageArgs map[string]interface{}

var idCounter uint

type Message struct {
	Command string      // message type
	Args    messageArgs // message arguments
	ID      uint        // message ID
}

// creates a new Message with an automatically-generated ID
func NewMessage(cmd string, args messageArgs) Message {
	idCounter++
	return NewMessageWithID(cmd, args, idCounter)
}

// creates a new Message with the specified ID
func NewMessageWithID(cmd string, args messageArgs, id uint) Message {
	return Message{cmd, args, id}
}

// creates a new Message from JSON data
// this is ugly, but I don't think there's a nicer way to do it since we use
// a JSON array and not an object?
func MessageFromJson(data []byte) (msg Message, err error) {

	// decode
	var iface interface{}
	if err = json.Unmarshal(data, &iface); err != nil {
		return msg, err
	}

	// top level must be array
	ary, ok := iface.([]interface{})
	if !ok || len(ary) != 3 {
		err = errors.New("Message must be a JSON array of length 3")
		return
	}

	// first element must be command
	cmd, ok := ary[0].(string)
	if !ok || len(cmd) == 0 {
		err = errors.New("Message has no type")
		return
	}

	// second element must be object
	args, ok := ary[1].(map[string]interface{})
	if !ok {
		err = errors.New("Message content must be a JSON object")
		return
	}

	// third element must be integer command ID
	id, ok := ary[2].(float64)
	if !ok || id < 0 {
		err = errors.New("Message has no ID")
		return
	}

	return NewMessageWithID(cmd, args, uint(id)), nil
}

// fetch an arguments as a string
func (msg Message) String(arg string) string {
	iface, ok := msg.Args[arg]
	if !ok {
		return ""
	}
	switch val := iface.(type) {
	case nil:
		return ""
	case string:
		return val
	}
	return fmt.Sprintf("%v", iface)
}

// translates the Message to JSON
func (msg Message) ToJson() []byte {
	ary := [...]interface{}{msg.Command, msg.Args, msg.ID}
	json, _ := json.Marshal(ary)
	return json
}
