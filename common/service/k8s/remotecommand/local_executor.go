package remotecommand

import (
	"context"
	"github.com/creack/pty"
	"io"
	"k8s.io/client-go/tools/remotecommand"
	"log/slog"
	"os/exec"
)

type LocalExecutor struct {
	cmd *exec.Cmd
}

func NewLocalExecutor(cmd *exec.Cmd) *LocalExecutor {
	return &LocalExecutor{
		cmd: cmd,
	}
}

func (e LocalExecutor) StreamWithContext(ctx context.Context, options remotecommand.StreamOptions) error {
	tty, err := pty.Start(e.cmd)
	if err != nil {
		return err
	}
	defer func() {
		//加write是避免退出时，tty没有关闭
		tty.Write([]byte("exit"))
		err := tty.Close()
		if err != nil {
			slog.Error("tty close error", "err", err)
			return
		}
	}()

	if options.Tty && options.TerminalSizeQueue != nil {
		go func() {
			for {
				size := options.TerminalSizeQueue.Next()
				if size == nil {
					return
				}

				err := pty.Setsize(tty, &pty.Winsize{Rows: size.Height, Cols: size.Width})
				if err != nil {
					slog.Error("setsize error", "err", err)
				}
			}
		}()
	}
	go func() {
		for {
			buf := make([]byte, 32*1024)
			read, err := tty.Read(buf)
			if err != nil {
				slog.Error("tty read error", "err", err)
				return
			}

			_, err = options.Stdout.Write(buf[:read])
			if err != nil {
				slog.Error("stdout write error", "err", err)
				return
			}
		}
	}()
	go func() {
		_, err := io.Copy(tty, options.Stdin)
		if err != nil {
			slog.Error("stdin copy error", "err", err)
			return
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}
