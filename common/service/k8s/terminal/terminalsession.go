package terminal

import (
	"bytes"
	"context"
	"encoding/binary"
	"log/slog"
	"sync"
	"unicode/utf8"

	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

type TerminalSession struct {
	conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
	reader   *bytes.Buffer
	writer   *bytes.Buffer
	context  context.Context
	cancel   context.CancelFunc
	once     sync.Once
}

func NewTerminalSession(conn *websocket.Conn) *TerminalSession {
	sizeChan := make(chan remotecommand.TerminalSize, 1)
	ctx, cancel := context.WithCancel(context.Background())
	return &TerminalSession{
		sizeChan: sizeChan,
		context:  ctx,
		conn:     conn,
		reader:   &bytes.Buffer{},
		writer:   &bytes.Buffer{},
		cancel:   cancel,
		once:     sync.Once{},
	}
}

func (t *TerminalSession) Next() *remotecommand.TerminalSize {
	size := <-t.sizeChan
	if size.Height == 0 && size.Width == 0 {
		return nil
	}
	return &size
}

func (t *TerminalSession) Read(p []byte) (n int, err error) {
	if t.conn != nil {
		msgType, data, err := t.conn.ReadMessage()
		if err != nil {
			t.Close()
			return 0, err
		}
		if msgType == websocket.BinaryMessage {
			var Cols, Rows uint16
			if err := binary.Read(bytes.NewReader(data[:2]), binary.LittleEndian, &Rows); err != nil {
				return 0, nil
				// continue
			}
			if err := binary.Read(bytes.NewReader(data[2:]), binary.LittleEndian, &Cols); err != nil {
				return 0, nil
			}
			//打印Rows Cols
			t.sizeChan <- remotecommand.TerminalSize{Width: Cols, Height: Rows}
			return 0, nil
		}
		if msgType == websocket.TextMessage {
			n = copy(p, data)
			return n, nil
		}
		return 0, nil
	}
	return t.reader.Read(p)
}

func (t *TerminalSession) Write(p []byte) (n int, err error) {
	// log.Println((string)(p))
	if t.conn != nil && utf8.Valid(p) {
		err := t.conn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			t.Close()
			slog.Info("write conn err", "err", err)
			return 0, err
		}
		return len(p), err
	}
	return t.writer.Write(p)
}

func (t *TerminalSession) Close() {
	t.once.Do(func() {
		if t.conn != nil {
			t.conn.Close()
		}
		if t.cancel != nil {
			slog.Info("k8s exec close context done")
			t.cancel()
		}
		close(t.sizeChan)
	})
}

func (t *TerminalSession) Done() <-chan struct{} {
	return t.context.Done()
}

func (t *TerminalSession) Context() context.Context {
	return t.context
}

func (t *TerminalSession) GetWriterBytes() []byte {
	return t.writer.Bytes()
}

// func (t *TerminalSession) Resize(size remotecommand.TerminalSize) {
// 	log.Println("k8s exec resize")
// 	buf := make([]byte, 2)
// 	binary.LittleEndian.PutUint16(buf, uint16(size.Height))
// 	binary.LittleEndian.PutUint16(buf[2:], uint16(size.Width))
// 	t.conn.WriteMessage(websocket.BinaryMessage, buf)
// }
