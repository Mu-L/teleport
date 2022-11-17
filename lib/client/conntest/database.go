/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package conntest

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/gravitational/trace"

	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/client"
	"github.com/gravitational/teleport/lib/client/conntest/database"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/tlsca"
)

// DatabasePinger describes the required methods to test a Database Connection.
type DatabasePinger interface {

	// Ping tests the connection to the Database with a simple request.
	Ping(ctx context.Context, req database.PingRequest) error

	// IsConnectionRefusedError returns whether the error is referring to a con.
	IsConnectionRefusedError(error) bool

	// IsInvalidDatabaseUserError returns whether the error is referring to an invalid (non-existent) user.
	IsInvalidDatabaseUserError(error) bool

	// IsInvalidDatabaseNameError returns whether the error is referring to an invalid (non-existent) user.
	IsInvalidDatabaseNameError(error) bool
}

// ClientDatabaseConnectionTester contains the required auth.ClientI methods to test a Database Connection
type ClientDatabaseConnectionTester interface {
	client.ALPNAuthTunnel

	services.ConnectionsDiagnostic

	// ListResoures returns a paginated list of resources.
	ListResources(ctx context.Context, req proto.ListResourcesRequest) (*types.ListResourcesResponse, error)
}

// DatabaseConnectionTesterConfig defines the config fields for DatabaseConnectionTester.
type DatabaseConnectionTesterConfig struct {
	// UserClient is an auth client that has a User's identity.
	UserClient ClientDatabaseConnectionTester

	// ProxyHostPort is the proxy to use in the `--proxy` format (host:webPort,sshPort)
	ProxyHostPort string

	// TLSRoutingEnabled indicates that proxy supports ALPN SNI server where
	// all proxy services are exposed on a single TLS listener (Proxy Web Listener).
	TLSRoutingEnabled bool
}

// DatabaseConnectionTester implements the ConnectionTester interface for Testing Database access.
type DatabaseConnectionTester struct {
	cfg          DatabaseConnectionTesterConfig
	webProxyAddr string
}

// NewDatabaseConnectionTester returns a new DatabaseConnectionTester
func NewDatabaseConnectionTester(cfg DatabaseConnectionTesterConfig) (*DatabaseConnectionTester, error) {
	parsedProxyHostAddr, err := client.ParseProxyHost(cfg.ProxyHostPort)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &DatabaseConnectionTester{
		cfg:          cfg,
		webProxyAddr: parsedProxyHostAddr.WebProxyAddr,
	}, nil
}

// TestConnection tests the access to a database using:
// - auth Client using the User access
// - the resource name
// - database user and database name to connect to
//
// A new ConnectionDiagnostic is created and used to store the traces as it goes through the checkpoints
// To connect to the Database, we will create a cert-key pair and setup a Database client back to Teleport Proxy.
// The following checkpoints are reported:
// - database server for the requested database exists / the user's roles can access it
// - the user can use the requested database user and database name (per their roles)
// - the database is acessible and accepting connections from the database server
// - the database has the database user and database name that was requested
func (s *DatabaseConnectionTester) TestConnection(ctx context.Context, req TestConnectionRequest) (types.ConnectionDiagnostic, error) {
	if req.ResourceKind != types.KindDatabase {
		return nil, trace.BadParameter("invalid value for ResourceKind, expected %q got %q", types.KindDatabase, req.ResourceKind)
	}

	if req.DatabaseUser == "" {
		return nil, trace.BadParameter("missing required parameter Database User")
	}

	if req.DatabaseName == "" {
		return nil, trace.BadParameter("missing required parameter Database Name")
	}

	connectionDiagnosticID := uuid.NewString()
	connectionDiagnostic, err := types.NewConnectionDiagnosticV1(
		connectionDiagnosticID,
		map[string]string{},
		types.ConnectionDiagnosticSpecV1{
			// We start with a failed state so that we don't need to set it to each return statement once an error is returned.
			// if the test reaches the end, we force the test to be a success by calling
			// 	connectionDiagnostic.SetMessage(types.DiagnosticMessageSuccess)
			//	connectionDiagnostic.SetSuccess(true)
			Message: types.DiagnosticMessageFailed,
		},
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if err := s.cfg.UserClient.CreateConnectionDiagnostic(ctx, connectionDiagnostic); err != nil {
		return nil, trace.Wrap(err)
	}

	// Lookup the Database Server that's proxying the requested Database.
	listResourcesResponse, err := s.cfg.UserClient.ListResources(ctx, proto.ListResourcesRequest{
		PredicateExpression: fmt.Sprintf(`name == "%s"`, req.ResourceName),
		ResourceType:        types.KindDatabaseServer,
		Limit:               1,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	databaseServers, err := types.ResourcesWithLabels(listResourcesResponse.Resources).AsDatabaseServers()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if len(databaseServers) == 0 {
		connDiag, err := s.appendDiagnosticTrace(ctx,
			connectionDiagnosticID,
			types.ConnectionDiagnosticTrace_RBAC_DATABASE,
			"You are not authorized to access this Database. "+
				"Ensure your role grants access by adding it to the 'db_labels' property. "+
				"This can also happen when you don't have a Database Agent proxying the database - "+
				"you can fix that by adding the database labels to the 'db_service.resources.labels' in 'teleport.yaml' file of the database agent.",
			trace.NotFound("%s not found", req.ResourceName),
		)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		return connDiag, nil
	}

	databaseServer := databaseServers[0]
	databaseProtocol := databaseServer.GetDatabase().GetProtocol()
	databaseServiceName := databaseServer.GetName()

	var databasePinger DatabasePinger

	switch databaseProtocol {
	case defaults.ProtocolPostgres:
		databasePinger = database.PostgresPinger{}
	default:
		return nil, trace.Errorf("Database protocol not supported yet for Test Connection")
	}

	if _, err := s.appendDiagnosticTrace(ctx,
		connectionDiagnosticID,
		types.ConnectionDiagnosticTrace_RBAC_DATABASE,
		"A Database Agent is available to proxy the connection to the Database.",
		nil,
	); err != nil {
		return nil, trace.Wrap(err)
	}

	list, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, trace.Wrap(err)
	}

	proxyMiddleware := client.NewALPNCertChecker().
		WithIdentityCheckerFunc(func(identity *tlsca.Identity) error {
			if req.DatabaseUser != identity.RouteToDatabase.Username {
				msg := fmt.Sprintf("certificate subject is for user %s, but need %s", identity.RouteToDatabase.Username, req.DatabaseUser)
				return trace.Wrap(errors.New(msg))
			}

			if req.DatabaseName != identity.RouteToDatabase.Database {
				msg := fmt.Sprintf("certificate subject is for database name %s, but need %s", identity.RouteToDatabase.Database, req.DatabaseName)
				return trace.Wrap(errors.New(msg))
			}

			return nil
		})

	err = client.RunALPNAuthTunnel(ctx, client.RunALPNAuthTunnelRequest{
		Client:                 s.cfg.UserClient,
		Listener:               list,
		Protocol:               databaseProtocol,
		WebProxyAddr:           s.webProxyAddr,
		ProxyMiddleware:        proxyMiddleware,
		ConnectionDiagnosticID: connectionDiagnosticID,
		RouteToDatabase: proto.RouteToDatabase{
			Protocol:    databaseProtocol,
			Username:    req.DatabaseUser,
			Database:    req.DatabaseName,
			ServiceName: databaseServiceName,
		},
		Insecure: true, // TODO(marco): remove this
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	pingRequest, err := newPingRequest(list.Addr().String(), req.DatabaseUser, req.DatabaseName)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	if err := databasePinger.Ping(ctx, pingRequest); err != nil {
		return s.handleErrFromPing(ctx, connectionDiagnosticID, err, databasePinger)
	}

	return s.handlePingSuccess(ctx, connectionDiagnosticID, databasePinger)
}

func newPingRequest(alpnProxyAddr, databaseUser, databaseName string) (database.PingRequest, error) {
	proxyHost, proxyPortStr, err := net.SplitHostPort(alpnProxyAddr)
	if err != nil {
		return database.PingRequest{}, trace.Wrap(err)
	}

	proxyPort, err := strconv.Atoi(proxyPortStr)
	if err != nil {
		return database.PingRequest{}, trace.Wrap(err)
	}

	return database.PingRequest{
		Host:     proxyHost,
		Port:     proxyPort,
		Username: databaseUser,
		Database: databaseName,
	}, nil
}

func (s DatabaseConnectionTester) handlePingSuccess(ctx context.Context, connectionDiagnosticID string, databasePinger DatabasePinger) (types.ConnectionDiagnostic, error) {
	if _, err := s.appendDiagnosticTrace(ctx, connectionDiagnosticID,
		types.ConnectionDiagnosticTrace_CONNECTIVITY,
		"Database is accessible from the Database Agent.",
		nil,
	); err != nil {
		return nil, trace.Wrap(err)
	}

	if _, err := s.appendDiagnosticTrace(ctx, connectionDiagnosticID,
		types.ConnectionDiagnosticTrace_RBAC_DATABASE_LOGIN,
		"Access to Database User and Database Name granted.",
		nil,
	); err != nil {
		return nil, trace.Wrap(err)
	}

	if _, err := s.appendDiagnosticTrace(ctx, connectionDiagnosticID,
		types.ConnectionDiagnosticTrace_DATABASE_DB_USER,
		"Database User exists in the Database.",
		nil,
	); err != nil {
		return nil, trace.Wrap(err)
	}

	connDiag, err := s.appendDiagnosticTrace(ctx, connectionDiagnosticID,
		types.ConnectionDiagnosticTrace_DATABASE_DB_NAME,
		"Database Name exists in the Database.",
		nil,
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	connDiag.SetMessage(types.DiagnosticMessageSuccess)
	connDiag.SetSuccess(true)

	if err := s.cfg.UserClient.UpdateConnectionDiagnostic(ctx, connDiag); err != nil {
		return nil, trace.Wrap(err)
	}

	return connDiag, nil
}

func (s DatabaseConnectionTester) handleErrFromPing(ctx context.Context, connectionDiagnosticID string, pingErr error, databasePinger DatabasePinger) (types.ConnectionDiagnostic, error) {
	// If the requested DB User/Name can't be used per RBAC checks, the Database Agent returns an error which gets here.
	// It must be ignored because there's already a Connection Diagnostic Trace written by the Database Agent (lib/srv/db/server.go)
	if strings.Contains(pingErr.Error(), "access to db denied. User does not have permissions. Confirm database user and name.") {
		connDiag, err := s.cfg.UserClient.GetConnectionDiagnostic(ctx, connectionDiagnosticID)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		return connDiag, nil
	}

	if databasePinger.IsConnectionRefusedError(pingErr) {
		connDiag, err := s.appendDiagnosticTrace(ctx,
			connectionDiagnosticID,
			types.ConnectionDiagnosticTrace_CONNECTIVITY,
			"There was a connection problem between the Database Agent and the Database. "+
				"Ensure the Database is running and accessible from the Database Agent.",
			pingErr,
		)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		return connDiag, nil
	}

	// Requested DB User is allowed per RBAC rules, but those entities don't exist in the Database itself.
	if databasePinger.IsInvalidDatabaseUserError(pingErr) {
		connDiag, err := s.appendDiagnosticTrace(ctx,
			connectionDiagnosticID,
			types.ConnectionDiagnosticTrace_DATABASE_DB_USER,
			"The Database rejected the provided Database User. Ensure that the database user exists.",
			pingErr,
		)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		return connDiag, nil
	}

	// Requested DB Name is allowed per RBAC rules, but those entities don't exist in the Database itself.
	if databasePinger.IsInvalidDatabaseNameError(pingErr) {
		connDiag, err := s.appendDiagnosticTrace(ctx,
			connectionDiagnosticID,
			types.ConnectionDiagnosticTrace_DATABASE_DB_NAME,
			"The Database rejected the provided Database Name. Ensure that the database name exists.",
			pingErr,
		)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		return connDiag, nil
	}

	connDiag, err := s.appendDiagnosticTrace(ctx,
		connectionDiagnosticID,
		types.ConnectionDiagnosticTrace_UNKNOWN_ERROR,
		"An unknown error occurred.",
		pingErr,
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return connDiag, nil
}

func (s DatabaseConnectionTester) appendDiagnosticTrace(ctx context.Context, connectionDiagnosticID string, traceType types.ConnectionDiagnosticTrace_TraceType, message string, err error) (types.ConnectionDiagnostic, error) {
	connDiag, err := s.cfg.UserClient.AppendDiagnosticTrace(
		ctx,
		connectionDiagnosticID,
		types.NewTraceDiagnosticConnection(
			traceType,
			message,
			err,
		))
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return connDiag, nil
}
