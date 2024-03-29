package jsonmsg

import (
	"encoding/json"
	"errors"
	"github.com/yinhm/ninepool/birpc"
	"io"
	"sync"
)

type codec struct {
	dec     *json.Decoder
	sending sync.Mutex
	enc     *json.Encoder
	closer  io.Closer
}

// This is ugly, but i need to override the unmarshaling logic for
// Args and Result, or they'll end up as map[string]interface{}.
// Perhaps some day encoding/json will support embedded structs, and I
// can embed birpc.Message and just override the two fields I need to
// change.
type jsonMessage struct {
	ID     uint64           `json:"id,omitempty"`
	Func   string           `json:"method,omitempty"`
	Args   json.RawMessage  `json:"params,omitempty"`
	Result json.RawMessage  `json:"result,omitempty"`
	Error  *json.RawMessage `json:"error,omitempty"`
}

type responseMessage struct {
	ID     uint64      `json:"id,omitempty"`
	Func   string      `json:"method,omitempty"`
	Args   interface{} `json:"params,omitempty"`
	Result interface{} `json:"result,omitempty"`
	Error  interface{} `json:"error,omitempty"`
}

type Notification struct {
	Func   string      `json:"method,omitempty"`
	Args   interface{} `json:"params,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

func (c *codec) ReadMessage(msg *birpc.Message) error {
	var jm jsonMessage
	err := c.dec.Decode(&jm)
	if err != nil {
		return err
	}
	msg.ID = jm.ID
	msg.Func = jm.Func
	msg.Args = jm.Args
	msg.Result = jm.Result

	if jm.Error != nil {
		rerr := &birpc.Error{}
		err = c.UnmarshalError(jm.Error, rerr)
		if err != nil {
			return err
		}
		msg.Error = rerr
	}

	return nil
}

func (c *codec) WriteMessage(msg *birpc.Message) error {
	c.sending.Lock()
	defer c.sending.Unlock()

	// notifcation hack
	n := Notification{}
	if msg.ID == 0 {
		n.Func = msg.Func
		n.Args = msg.Args
		n.Result = msg.Result
		return c.enc.Encode(n)
	}

	var r responseMessage
	r.ID = msg.ID
	r.Func = msg.Func
	r.Args = msg.Args
	r.Args = msg.Args
	r.Result = msg.Result
	if msg.Error != nil {
		r.Error = &birpc.List{
			msg.Error.Code,
			msg.Error.Msg,
			msg.Error.Data,
		}
	}
	return c.enc.Encode(r)
}

func (c *codec) Close() error {
	return c.closer.Close()
}

func (c *codec) UnmarshalArgs(msg *birpc.Message, args interface{}) error {
	raw := msg.Args.(json.RawMessage)
	if raw == nil {
		return nil
	}
	err := json.Unmarshal(raw, args)
	return err
}

func (c *codec) UnmarshalResult(msg *birpc.Message, result interface{}) error {
	raw := msg.Result.(json.RawMessage)
	if raw == nil {
		return errors.New("birpc.jsonmsg response must set result")
	}
	err := json.Unmarshal(raw, result)
	return err
}

func (c *codec) UnmarshalError(raw *json.RawMessage, rerr *birpc.Error) error {
	if raw == nil {
		return nil
	}
	to := &birpc.List{}
	err := json.Unmarshal([]byte(*raw), to)
	if err != nil {
		return err
	}
	d := (birpc.List)(*to)

	rerr.Code = int(d[0].(float64))
	rerr.Msg = d[1].(string)
	if len(d) == 3 {
		rerr.Data = d[2]
	}
	return nil
}

func NewCodec(conn io.ReadWriteCloser) *codec {
	c := &codec{
		dec:    json.NewDecoder(conn),
		enc:    json.NewEncoder(conn),
		closer: conn,
	}
	return c
}
