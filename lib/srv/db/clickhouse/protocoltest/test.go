/*
 * Teleport
 * Copyright (C) 2023  Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package protocoltest

import (
	"context"
	"crypto/tls"
	"database/sql"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"

	"github.com/ClickHouse/ch-go"
	chproto "github.com/ClickHouse/ch-go/proto"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/proto"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/srv/db/common"
	"github.com/gravitational/teleport/lib/utils/log/logtest"
	sliceutils "github.com/gravitational/teleport/lib/utils/slices"
)

// TestServerOption allows setting test server options.
type TestServerOption func(*TestServer)

// WithClickHouseNativeProtocol specify the Native Protocol for test TestServer.
func WithClickHouseNativeProtocol() TestServerOption {
	return func(server *TestServer) {
		server.protocol = defaults.ProtocolClickHouse
	}
}

// WithClickHouseHTTPProtocol specify the HTTP ClickHouse Protocol for test TestServer.
func WithClickHouseHTTPProtocol() TestServerOption {
	return func(server *TestServer) {
		server.protocol = defaults.ProtocolClickHouseHTTP
	}
}

// TestServer is a ClickHouse test server that allows to handle
// basic HTTP and ClickHouse Native connections.
type TestServer struct {
	cfg       common.TestServerConfig
	listener  net.Listener
	port      string
	tlsConfig *tls.Config
	logger    *slog.Logger
	protocol  string
}

// NewTestServer returns a new instance of a test ClickHouse server.
func NewTestServer(config common.TestServerConfig, opts ...TestServerOption) (*TestServer, error) {
	err := config.CheckAndSetDefaults()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	tlsConfig, err := common.MakeTestServerTLSConfig(config)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	tlsConfig.InsecureSkipVerify = true
	if config.Listener != nil {
		config.Listener = tls.NewListener(config.Listener, tlsConfig)
	}
	port, err := config.Port()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	testServer := &TestServer{
		cfg:       config,
		listener:  config.Listener,
		port:      port,
		tlsConfig: tlsConfig,
		logger: logtest.With(
			teleport.ComponentKey, defaults.ProtocolClickHouse,
			"name", config.Name,
		),
	}

	for _, opt := range opts {
		opt(testServer)
	}

	return testServer, nil
}

func (s *TestServer) Serve() error {
	switch s.protocol {
	case defaults.ProtocolClickHouseHTTP:
		return trace.Wrap(s.serveHTTP())
	default:
		return trace.Wrap(s.serveNative())
	}
}

func encodeHello() ([]byte, error) {
	var block proto.Block

	type columnWithValue struct {
		name       string
		columnType column.Type
		value      any
	}
	columns := []columnWithValue{
		{
			name:       "displayName()",
			columnType: "String",
			value:      "ClickHouse",
		},
		{
			name:       "version()",
			columnType: "String",
			// x509 HTTP auth is support from ClickHouse 22.4.x.x
			// Report a random version to a ClickHouse HTTP client.
			value: "23.4.2.11",
		},
		{
			name:       "revision()",
			columnType: "UInt32",
			value:      uint32(12345),
		},
		{
			name:       "timezone()",
			columnType: "String",
			value:      "UTC",
		},
	}

	for _, c := range columns {
		if err := block.AddColumn(c.name, c.columnType); err != nil {
			return nil, trace.Wrap(err)
		}
	}

	values := sliceutils.Map(columns, func(c columnWithValue) any {
		return c.value
	})
	if err := block.Append(values...); err != nil {
		return nil, trace.Wrap(err)
	}

	var chb chproto.Buffer
	if err := block.Encode(&chb, 0); err != nil {
		return nil, trace.Wrap(err)
	}
	return chb.Buf, nil

}

func encodePing() ([]byte, error) {
	block := proto.Block{}
	if err := block.AddColumn("1", "UInt8"); err != nil {
		return nil, trace.Wrap(err)
	}
	if err := block.Append(uint8(0)); err != nil {
		return nil, trace.Wrap(err)
	}
	var chb chproto.Buffer
	if err := block.Encode(&chb, 0); err != nil {
		return nil, trace.Wrap(err)
	}
	return chb.Buf, nil

}

const (
	// HelloQuery is the "hello" query sent by the HTTP client to select a few
	// common metadata.
	// https://github.com/ClickHouse/clickhouse-go/blob/10732d7bb20224020e7099e9675f4c47ae5f5e7f/conn_http.go#L321-L324
	HelloQuery = "SELECT displayName(), version(), revision(), timezone()"
	// PingQuery is the basic query to ping.
	PingQuery = "SELECT 1"
)

func (s *TestServer) serveHTTP() error {
	encHandler := map[string]func() ([]byte, error){
		HelloQuery: encodeHello,
		PingQuery:  encodePing,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		buff, err := io.ReadAll(request.Body)
		if err != nil {
			s.logger.ErrorContext(request.Context(), "Got unexpected error", "error", err)
		}
		defer request.Body.Close()

		query := string(buff)
		enc, ok := encHandler[query]
		if !ok {
			s.logger.ErrorContext(request.Context(), "Got unexpected query", "query", query)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		respBuff, err := enc()
		if err != nil {
			s.logger.ErrorContext(request.Context(), "Got unexpected error", "error", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(respBuff)
		if err != nil {
			s.logger.ErrorContext(request.Context(), "Got unexpected error", "error", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	srv := &httptest.Server{
		Listener: s.listener,
		Config:   &http.Server{Handler: mux},
	}

	srv.Start()
	return nil
}

func (s *TestServer) serveNative() error {
	server := ch.NewServer(ch.ServerOptions{})
	if err := server.Serve(s.listener); err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// Port returns Clickhouse Server port.
func (s *TestServer) Port() string {
	return s.port
}

// Close closes the server listener.
func (s *TestServer) Close() error {
	return s.listener.Close()
}

// MakeNativeTestClient returns ClickHouse Native Server client used in tests.
func MakeNativeTestClient(ctx context.Context, config common.TestClientConfig) (*ch.Client, error) {
	conn, err := ch.Dial(ctx, ch.Options{
		Address: config.Address,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return conn, nil
}

// MakeDBTestClient returns ClickHouse HTTP Server client used in tests.
func MakeDBTestClient(ctx context.Context, config common.TestClientConfig) (*sql.DB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Protocol: toClickhouseProtocol(config.RouteToDatabase.Protocol),
		Addr:     []string{config.Address},
	})
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, trace.Wrap(err)
	}
	return conn, nil
}

func toClickhouseProtocol(protocol string) clickhouse.Protocol {
	switch protocol {
	case defaults.ProtocolClickHouseHTTP:
		return clickhouse.HTTP
	}
	return clickhouse.Native
}
