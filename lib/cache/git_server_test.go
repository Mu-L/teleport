/*
 * Teleport
 * Copyright (C) 2024  Gravitational, Inc.
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

package cache

import (
	"context"
	"testing"

	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/api/utils/clientutils"
	"github.com/gravitational/teleport/lib/itertools/stream"
)

func TestGitServers(t *testing.T) {
	t.Parallel()

	p, err := newPack(t.TempDir(), ForAuth)
	require.NoError(t, err)
	t.Cleanup(p.Close)

	testResources(t, p, testFuncs[types.Server]{
		newResource: func(name string) (types.Server, error) {
			return types.NewGitHubServerWithName(name,
				types.GitHubServerMetadata{
					Integration:  name,
					Organization: name,
				})
		},
		create: func(ctx context.Context, server types.Server) error {
			_, err := p.gitServers.CreateGitServer(ctx, server)
			return trace.Wrap(err)
		},
		list: func(ctx context.Context) ([]types.Server, error) {
			return stream.Collect(clientutils.Resources(ctx, p.gitServers.ListGitServers))
		},
		update: func(ctx context.Context, server types.Server) error {
			_, err := p.gitServers.UpdateGitServer(ctx, server)
			return trace.Wrap(err)
		},
		deleteAll: p.gitServers.DeleteAllGitServers,
		cacheList: func(ctx context.Context, pageSize int) ([]types.Server, error) {
			return stream.Collect(clientutils.Resources(ctx, p.cache.ListGitServers))
		},
		cacheGet: p.cache.GetGitServer,
	})
}
