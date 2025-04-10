package ws

import (
	"bytes"
	"compress/flate"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/gobwas/httphead"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsflate"
	"github.com/gobwas/ws/wsutil"

	"realy.mleku.dev/context"
)

// Connection is an outbound client -> relay connection.
type Connection struct {
	conn              net.Conn
	enableCompression bool
	controlHandler    wsutil.FrameHandlerFunc
	flateReader       *wsflate.Reader
	reader            *wsutil.Reader
	flateWriter       *wsflate.Writer
	writer            *wsutil.Writer
	msgStateR         *wsflate.MessageState
	msgStateW         *wsflate.MessageState
}

// NewConnection creates a new Connection.
func NewConnection(c context.T, url string, requestHeader http.Header,
	tlsConfig *tls.Config) (*Connection, error) {
	dialer := ws.Dialer{
		Header: ws.HandshakeHeaderHTTP(requestHeader),
		Extensions: []httphead.Option{
			wsflate.DefaultParameters.Option(),
		},
		TLSConfig: tlsConfig,
	}
	conn, _, hs, err := dialer.Dial(c, url)
	if err != nil {
		return nil, errorf.E("failed to dial: %w", err)
	}

	enableCompression := false
	state := ws.StateClientSide
	for _, extension := range hs.Extensions {
		if string(extension.Name) == wsflate.ExtensionName {
			enableCompression = true
			state |= ws.StateExtended
			break
		}
	}

	// reader
	var flateReader *wsflate.Reader
	var msgStateR wsflate.MessageState
	if enableCompression {
		msgStateR.SetCompressed(true)

		flateReader = wsflate.NewReader(nil, func(r io.Reader) wsflate.Decompressor {
			return flate.NewReader(r)
		})
	}

	controlHandler := wsutil.ControlFrameHandler(conn, ws.StateClientSide)
	reader := &wsutil.Reader{
		Source:         conn,
		State:          state,
		OnIntermediate: controlHandler,
		CheckUTF8:      false,
		Extensions: []wsutil.RecvExtension{
			&msgStateR,
		},
	}

	// writer
	var flateWriter *wsflate.Writer
	var msgStateW wsflate.MessageState
	if enableCompression {
		msgStateW.SetCompressed(true)

		flateWriter = wsflate.NewWriter(nil, func(w io.Writer) wsflate.Compressor {
			fw, err := flate.NewWriter(w, 4)
			if err != nil {
				log.E.F("Failed to create flate writer: %v", err)
			}
			return fw
		})
	}

	writer := wsutil.NewWriter(conn, state, ws.OpText)
	writer.SetExtensions(&msgStateW)

	return &Connection{
		conn:              conn,
		enableCompression: enableCompression,
		controlHandler:    controlHandler,
		flateReader:       flateReader,
		reader:            reader,
		msgStateR:         &msgStateR,
		flateWriter:       flateWriter,
		writer:            writer,
		msgStateW:         &msgStateW,
	}, nil
}

// WriteMessage dispatches a message through the Connection.
func (cn *Connection) WriteMessage(c context.T, data []byte) error {
	select {
	case <-c.Done():
		return errors.New("context canceled")
	default:
	}

	if cn.msgStateW.IsCompressed() && cn.enableCompression {
		cn.flateWriter.Reset(cn.writer)
		if _, err := io.Copy(cn.flateWriter, bytes.NewReader(data)); chk.T(err) {
			return errorf.E("failed to write message: %w", err)
		}

		if err := cn.flateWriter.Close(); chk.T(err) {
			return errorf.E("failed to close flate writer: %w", err)
		}
	} else {
		if _, err := io.Copy(cn.writer, bytes.NewReader(data)); chk.T(err) {
			return errorf.E("failed to write message: %w", err)
		}
	}

	if err := cn.writer.Flush(); chk.T(err) {
		return errorf.E("failed to flush writer: %w", err)
	}

	return nil
}

// ReadMessage picks up the next incoming message on a Connection.
func (cn *Connection) ReadMessage(c context.T, buf io.Writer) error {
	for {
		select {
		case <-c.Done():
			return errors.New("context canceled")
		default:
		}

		h, err := cn.reader.NextFrame()
		if err != nil {
			cn.conn.Close()
			return errorf.E("failed to advance frame: %w", err)
		}

		if h.OpCode.IsControl() {
			if err := cn.controlHandler(h, cn.reader); chk.T(err) {
				return errorf.E("failed to handle control frame: %w", err)
			}
		} else if h.OpCode == ws.OpBinary ||
			h.OpCode == ws.OpText {
			break
		}

		if err := cn.reader.Discard(); chk.T(err) {
			return errorf.E("failed to discard: %w", err)
		}
	}

	if cn.msgStateR.IsCompressed() && cn.enableCompression {
		cn.flateReader.Reset(cn.reader)
		if _, err := io.Copy(buf, cn.flateReader); chk.T(err) {
			return errorf.E("failed to read message: %w", err)
		}
	} else {
		if _, err := io.Copy(buf, cn.reader); chk.T(err) {
			return errorf.E("failed to read message: %w", err)
		}
	}

	return nil
}

// Close the Connection.
func (cn *Connection) Close() error {
	return cn.conn.Close()
}
