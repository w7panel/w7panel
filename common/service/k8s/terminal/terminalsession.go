package terminal

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
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
	mu       sync.Mutex
	closedBy string
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
		closedBy: "unknown",
	}
}

func closeReasonFromWsError(err error) string {
	if err == nil {
		return "unknown"
	}
	if errors.Is(err, io.EOF) {
		return "client_close"
	}
	var closeErr *websocket.CloseError
	if errors.As(err, &closeErr) {
		switch closeErr.Code {
		case websocket.CloseNormalClosure:
			return "client_close"
		case websocket.CloseGoingAway:
			return "upstream_close"
		case websocket.CloseAbnormalClosure:
			return "upstream_close"
		default:
			return "upstream_close"
		}
	}
	return "upstream_close"
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
			reason := closeReasonFromWsError(err)
			slog.Info("websocket ReadMessage error, signaling EOF to stdin", "err", err, "reason", reason)
			t.CloseWithReason(reason)
			return 0, io.EOF
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
	if t.conn != nil && utf8.Valid(p) {
		err := t.conn.WriteMessage(websocket.TextMessage, p)
		if err != nil {
			reason := closeReasonFromWsError(err)
			slog.Info("write conn err", "err", err, "reason", reason)
			t.CloseWithReason(reason)
			return 0, io.EOF
		}
		return len(p), err
	}
	return t.writer.Write(p)
}

func (t *TerminalSession) Close() {
	t.CloseWithReason("session_close")
}

func (t *TerminalSession) CloseWithReason(reason string) {
	if reason == "" {
		reason = "unknown"
	}
	t.mu.Lock()
	t.closedBy = reason
	t.mu.Unlock()

	t.once.Do(func() {
		reasonSnapshot := t.GetCloseReason()
		if t.context != nil {
			if errors.Is(t.context.Err(), context.DeadlineExceeded) {
				reasonSnapshot = "timeout"
			}
		}
		if t.cancel != nil {
			slog.Info("k8s exec close context done", "reason", reasonSnapshot)
			t.cancel()
		}
		if t.conn != nil {
			t.conn.Close()
		}
		close(t.sizeChan)
	})
}

func (t *TerminalSession) GetCloseReason() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closedBy == "" {
		return "unknown"
	}
	return t.closedBy
}

func (t *TerminalSession) Done() <-chan struct{} {
	return t.context.Done()
}

func (t *TerminalSession) Context() context.Context {
	return t.context
}

func (t *TerminalSession) SetContext(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	t.cancel = cancel
	t.context = ctx
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
