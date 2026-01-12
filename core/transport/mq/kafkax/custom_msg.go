package kafkax

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type RawMessage map[string]any

func NewRawMessage() RawMessage {
	return make(map[string]any)
}

func (m *RawMessage) RawMessageInject(ctx context.Context) {
	carrire := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrire)
}

func (m *RawMessage) Marshal() []byte {
	bts, _ := json.Marshal(m)
	return bts
}

type CustomMessage struct {
	Header  RawMessage `json:"header"`
	Payload RawMessage `json:"payload"`
}

func NewMessage(ms ...map[string]any) *CustomMessage {
	if len(ms) <= 0 {
		return &CustomMessage{
			Header:  make(map[string]any),
			Payload: make(map[string]any),
		}
	}
	if len(ms) == 1 {
		return &CustomMessage{
			Header:  make(map[string]any),
			Payload: ms[0],
		}
	}
	return &CustomMessage{
		Header:  ms[1],
		Payload: ms[0],
	}
}

func (m *CustomMessage) NewPubMsg(key []byte) *PubMsg {
	bs, _ := json.Marshal(m)
	return &PubMsg{
		key:   key,
		value: bs,
	}
}

type PubMsg struct {
	key   []byte // 消息key
	value []byte // 消息内容
}

type PubResult struct {
	err    error
	srcMsg *PubMsg
}

func MQMsgBuilder(key, value []byte) *PubMsg {
	return &PubMsg{
		key:   key,
		value: value,
	}
}

func (pm *PubMsg) Key() []byte {
	return pm.key
}

func (pm *PubMsg) Value() []byte {
	return pm.value
}

func (pr *PubResult) Error() error {
	return pr.err
}

func (pr *PubResult) SrcMsg() *PubMsg {
	return pr.srcMsg
}

func (pr *PubResult) SetError(err error) {
	pr.err = err
}

func (pr *PubResult) SetSrcMsg(key, value []byte) {
	pr.srcMsg = &PubMsg{key: key, value: value}
}
