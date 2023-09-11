package messages

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// CommandCompleteTag indicates which SQL command was completed.
type CommandCompleteTag byte

const (
	CommandCompleteTag_INSERT CommandCompleteTag = iota
	CommandCompleteTag_DELETE
	CommandCompleteTag_UPDATE
	CommandCompleteTag_MERGE
	CommandCompleteTag_SELECT
	CommandCompleteTag_MOVE
	CommandCompleteTag_FETCH
	CommandCompleteTag_COPY
)

// CommandComplete tells the client that the command has completed.
type CommandComplete struct {
	Tag  CommandCompleteTag
	Rows int32
}

// Bytes returns CommandComplete as a byte slice, ready to be returned to the client.
func (cc CommandComplete) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('C')          // Message Type
	WriteNumber(&buf, int32(0)) // Message length, will be corrected later
	switch cc.Tag {
	case CommandCompleteTag_INSERT:
		buf.WriteString("INSERT 0 ")
	case CommandCompleteTag_DELETE:
		buf.WriteString("DELETE ")
	case CommandCompleteTag_UPDATE:
		buf.WriteString("UPDATE ")
	case CommandCompleteTag_MERGE:
		buf.WriteString("MERGE ")
	case CommandCompleteTag_SELECT:
		buf.WriteString("SELECT ")
	case CommandCompleteTag_MOVE:
		buf.WriteString("MOVE ")
	case CommandCompleteTag_FETCH:
		buf.WriteString("FETCH ")
	case CommandCompleteTag_COPY:
		buf.WriteString("COPY ")
	}
	buf.WriteString(strconv.Itoa(int(cc.Rows)))
	buf.WriteByte(0) // Trailing NULL character, denoting the end of the string
	return WriteLength(buf.Bytes())
}

// QueryToCommandCompleteTag returns the appropriate command tag for the given query.
func QueryToCommandCompleteTag(query string) (CommandCompleteTag, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if strings.HasPrefix(query, "select") {
		return CommandCompleteTag_SELECT, nil
	} else if strings.HasPrefix(query, "insert") {
		return CommandCompleteTag_INSERT, nil
	} else if strings.HasPrefix(query, "update") {
		return CommandCompleteTag_UPDATE, nil
	} else if strings.HasPrefix(query, "delete") {
		return CommandCompleteTag_DELETE, nil
	} else if strings.HasPrefix(query, "create") {
		return CommandCompleteTag_SELECT, nil
	} else if strings.HasPrefix(query, "call") {
		return CommandCompleteTag_SELECT, nil
	} else {
		return 0, fmt.Errorf("unsupported query for now")
	}
}
