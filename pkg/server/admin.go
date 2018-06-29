// Copyright 2014 The Cockroach Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cockroachdb/apd"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/internal/client"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/debug"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const (
	// adminPrefix is the prefix for RESTful endpoints used to provide an
	// administrative interface to the cockroach cluster.
	adminPrefix = "/_admin/v1/"

	// defaultAPIEventLimit is the default maximum number of events returned by any
	// endpoints returning events.
	defaultAPIEventLimit = 1000
)

// apiServerMessage is the standard body for all HTTP 500 responses.
var errAdminAPIError = status.Errorf(codes.Internal, "An internal server error "+
	"has occurred. Please check your CockroachDB logs for more details.")

// A adminServer provides a RESTful HTTP API to administration of
// the cockroach cluster.
type adminServer struct {
	server     *Server
	memMonitor mon.BytesMonitor
	memMetrics *sql.MemoryMetrics
}

// noteworthyAdminMemoryUsageBytes is the minimum size tracked by the
// admin SQL pool before the pool start explicitly logging overall
// usage growth in the log.
var noteworthyAdminMemoryUsageBytes = envutil.EnvOrDefaultInt64("COCKROACH_NOTEWORTHY_ADMIN_MEMORY_USAGE", 100*1024)

// newAdminServer allocates and returns a new REST server for
// administrative APIs.
func newAdminServer(s *Server) *adminServer {
	server := &adminServer{server: s, memMetrics: &s.adminMemMetrics}
	// TODO(knz): We do not limit memory usage by admin operations
	// yet. Is this wise?
	server.memMonitor = mon.MakeUnlimitedMonitor(
		context.Background(),
		"admin",
		mon.MemoryResource,
		nil,
		nil,
		noteworthyAdminMemoryUsageBytes,
		s.ClusterSettings(),
	)
	return server
}

// RegisterService registers the GRPC service.
func (s *adminServer) RegisterService(g *grpc.Server) {
	serverpb.RegisterAdminServer(g, s)
}

// RegisterGateway starts the gateway (i.e. reverse proxy) that proxies HTTP requests
// to the appropriate gRPC endpoints.
func (s *adminServer) RegisterGateway(
	ctx context.Context, mux *gwruntime.ServeMux, conn *grpc.ClientConn,
) error {
	return serverpb.RegisterAdminHandler(ctx, mux, conn)
}

// getUserProto will return the authenticated user. For now, this is just a stub until we
// figure out our authentication mechanism.
//
// TODO(cdo): Make this work when we have an authentication scheme for the
// API.
func (s *adminServer) getUser(_ protoutil.Message) string {
	return security.RootUser
}

// serverError logs the provided error and returns an error that should be returned by
// the RPC endpoint method.
func (s *adminServer) serverError(err error) error {
	log.ErrorfDepth(context.TODO(), 1, "%s", err)
	return errAdminAPIError
}

// serverErrorf logs the provided error and returns an error that should be returned by
// the RPC endpoint method.
func (s *adminServer) serverErrorf(format string, args ...interface{}) error {
	log.ErrorfDepth(context.TODO(), 1, format, args...)
	return errAdminAPIError
}

// serverErrors logs the provided errors and returns an error that should be returned by
// the RPC endpoint method.
func (s *adminServer) serverErrors(errors []error) error {
	log.ErrorfDepth(context.TODO(), 1, "%v", errors)
	return errAdminAPIError
}

// isNotFoundError returns true if err is a table/database not found error.
func (s *adminServer) isNotFoundError(err error) bool {
	// TODO(cdo): Replace this crude suffix-matching with something more structured once we have
	// more structured errors.
	return err != nil && strings.HasSuffix(err.Error(), "does not exist")
}

// AllMetricMetadata returns all metrics' metadata.
func (s *adminServer) AllMetricMetadata(
	ctx context.Context, req *serverpb.MetricMetadataRequest,
) (*serverpb.MetricMetadataResponse, error) {

	resp := &serverpb.MetricMetadataResponse{
		Metadata: s.server.recorder.GetMetricsMetadata(),
	}

	return resp, nil
}

// Databases is an endpoint that returns a list of databases.
func (s *adminServer) Databases(
	ctx context.Context, req *serverpb.DatabasesRequest,
) (*serverpb.DatabasesResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}
	rows, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admi-show-db", nil /* txn */, args, "SHOW DATABASES",
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	var resp serverpb.DatabasesResponse
	for _, row := range rows {
		dbDatum, ok := tree.AsDString(row[0])
		if !ok {
			return nil, s.serverErrorf("type assertion failed on db name: %T", row[0])
		}
		dbName := string(dbDatum)
		resp.Databases = append(resp.Databases, dbName)
	}

	return &resp, nil
}

// DatabaseDetails is an endpoint that returns grants and a list of table names
// for the specified database.
func (s *adminServer) DatabaseDetails(
	ctx context.Context, req *serverpb.DatabaseDetailsRequest,
) (*serverpb.DatabaseDetailsResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}

	escDBName := tree.NameStringP(&req.Database)
	// Placeholders don't work with SHOW statements, so we need to manually
	// escape the database name.
	//
	// TODO(cdo): Use placeholders when they're supported by SHOW.

	// Marshal grants.
	rows, cols, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-show-grants", nil /* txn */, args,
		fmt.Sprintf("SHOW GRANTS ON DATABASE %s", escDBName),
	)
	if s.isNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "%s", err)
	}
	if err != nil {
		return nil, s.serverError(err)
	}
	var resp serverpb.DatabaseDetailsResponse
	{
		const (
			schemaCol     = "Schema"
			userCol       = "User"
			privilegesCol = "Privileges"
		)

		scanner := makeResultScanner(cols)
		for _, row := range rows {
			var schemaName string
			if err := scanner.Scan(row, schemaCol, &schemaName); err != nil {
				return nil, err
			}
			if schemaName != tree.PublicSchema {
				// We only want to list real tables.
				continue
			}

			// Marshal grant, splitting comma-separated privileges into a proper slice.
			var grant serverpb.DatabaseDetailsResponse_Grant
			var privileges string
			if err := scanner.Scan(row, userCol, &grant.User); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, privilegesCol, &privileges); err != nil {
				return nil, err
			}
			grant.Privileges = strings.Split(privileges, ",")
			resp.Grants = append(resp.Grants, grant)
		}
	}

	// Marshal table names.
	rows, cols, err = s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-show-tables", nil /* txn */, args,
		fmt.Sprintf("SHOW TABLES FROM %s", escDBName),
	)
	if s.isNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "%s", err)
	}
	if err != nil {
		return nil, s.serverError(err)
	}

	// Marshal table names.
	{
		const tableCol = "Table"
		scanner := makeResultScanner(cols)
		if a, e := len(cols), 1; a != e {
			return nil, s.serverErrorf("show tables columns mismatch: %d != expected %d", a, e)
		}
		for _, row := range rows {
			var tableName string
			if err := scanner.Scan(row, tableCol, &tableName); err != nil {
				return nil, err
			}
			resp.TableNames = append(resp.TableNames, tableName)
		}
	}

	// Query the descriptor ID and zone configuration for this database.
	{
		path, err := s.queryDescriptorIDPath(ctx, args, []string{req.Database})
		if err != nil {
			return nil, s.serverError(err)
		}
		resp.DescriptorID = int64(path[1])

		id, zone, zoneExists, err := s.queryZonePath(ctx, args, path)
		if err != nil {
			return nil, s.serverError(err)
		}

		if !zoneExists {
			zone = config.DefaultZoneConfig()
		}
		resp.ZoneConfig = zone

		switch id {
		case path[1]:
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_DATABASE
		default:
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_CLUSTER
		}
	}

	return &resp, nil
}

// TableDetails is an endpoint that returns columns, indices, and other
// relevant details for the specified table.
func (s *adminServer) TableDetails(
	ctx context.Context, req *serverpb.TableDetailsRequest,
) (*serverpb.TableDetailsResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}

	escDBName := tree.NameStringP(&req.Database)
	// TODO(cdo): Use real placeholders for the table and database names when we've extended our SQL
	// grammar to allow that.
	escTableName := tree.NameStringP(&req.Table)
	escQualTable := fmt.Sprintf("%s.%s", escDBName, escTableName)

	var resp serverpb.TableDetailsResponse

	// Marshal SHOW COLUMNS result.
	rows, cols, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-show-columns",
		nil /* txn */, args, fmt.Sprintf("SHOW COLUMNS FROM %s", escQualTable),
	)
	if s.isNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "%s", err)
	}
	if err != nil {
		return nil, s.serverError(err)
	}
	// TODO(cdo): protobuf v3's default behavior for fields with zero values (e.g. empty strings)
	// is to suppress them. So, if protobuf field "foo" is an empty string, "foo" won't show
	// up in the marshaled JSON. I feel that this is counterintuitive, and this should be fixed
	// for our API.
	{
		const (
			fieldCol   = "Field" // column name
			typeCol    = "Type"
			nullCol    = "Null"
			defaultCol = "Default"
		)
		scanner := makeResultScanner(cols)
		for _, row := range rows {
			var col serverpb.TableDetailsResponse_Column
			if err := scanner.Scan(row, fieldCol, &col.Name); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, typeCol, &col.Type); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, nullCol, &col.Nullable); err != nil {
				return nil, err
			}
			isDefaultNull, err := scanner.IsNull(row, defaultCol)
			if err != nil {
				return nil, err
			}
			if !isDefaultNull {
				if err := scanner.Scan(row, defaultCol, &col.DefaultValue); err != nil {
					return nil, err
				}
			}
			resp.Columns = append(resp.Columns, col)
		}
	}

	// Marshal SHOW INDEX result.
	rows, cols, err = s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-showindex",
		nil /* txn */, args, fmt.Sprintf("SHOW INDEX FROM %s", escQualTable),
	)
	if s.isNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "%s", err)
	}
	if err != nil {
		return nil, s.serverError(err)
	}
	{
		const (
			nameCol      = "Name"
			uniqueCol    = "Unique"
			seqCol       = "Seq"
			columnCol    = "Column"
			directionCol = "Direction"
			storingCol   = "Storing"
			implicitCol  = "Implicit"
		)
		scanner := makeResultScanner(cols)
		for _, row := range rows {
			// Marshal grant, splitting comma-separated privileges into a proper slice.
			var index serverpb.TableDetailsResponse_Index
			if err := scanner.Scan(row, nameCol, &index.Name); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, uniqueCol, &index.Unique); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, seqCol, &index.Seq); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, columnCol, &index.Column); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, directionCol, &index.Direction); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, storingCol, &index.Storing); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, implicitCol, &index.Implicit); err != nil {
				return nil, err
			}
			resp.Indexes = append(resp.Indexes, index)
		}
	}

	// Marshal SHOW GRANTS result.
	rows, cols, err = s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-show-grants",
		nil /* txn */, args, fmt.Sprintf("SHOW GRANTS ON TABLE %s", escQualTable),
	)
	if s.isNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "%s", err)
	}
	if err != nil {
		return nil, s.serverError(err)
	}
	{
		const (
			userCol       = "User"
			privilegesCol = "Privileges"
		)
		scanner := makeResultScanner(cols)
		for _, row := range rows {
			// Marshal grant, splitting comma-separated privileges into a proper slice.
			var grant serverpb.TableDetailsResponse_Grant
			var privileges string
			if err := scanner.Scan(row, userCol, &grant.User); err != nil {
				return nil, err
			}
			if err := scanner.Scan(row, privilegesCol, &privileges); err != nil {
				return nil, err
			}
			grant.Privileges = strings.Split(privileges, ",")
			resp.Grants = append(resp.Grants, grant)
		}
	}

	// Marshal SHOW CREATE TABLE result.
	rows, cols, err = s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-show-create",
		nil /* txn */, args, fmt.Sprintf("SHOW CREATE TABLE %s", escQualTable),
	)
	if s.isNotFoundError(err) {
		return nil, status.Errorf(codes.NotFound, "%s", err)
	}
	if err != nil {
		return nil, s.serverError(err)
	}
	{
		const createTableCol = "CreateTable"
		if len(rows) != 1 {
			return nil, s.serverErrorf("CreateTable response not available.")
		}

		scanner := makeResultScanner(cols)
		var createStmt string
		if err := scanner.Scan(rows[0], createTableCol, &createStmt); err != nil {
			return nil, err
		}

		resp.CreateTableStatement = createStmt
	}

	var tableID sqlbase.ID
	// Query the descriptor ID and zone configuration for this table.
	{
		path, err := s.queryDescriptorIDPath(ctx, args, []string{req.Database, req.Table})
		if err != nil {
			return nil, s.serverError(err)
		}
		tableID = path[2]
		resp.DescriptorID = int64(tableID)

		id, zone, zoneExists, err := s.queryZonePath(ctx, args, path)
		if err != nil {
			return nil, s.serverError(err)
		}

		if !zoneExists {
			zone = config.DefaultZoneConfig()
		}
		resp.ZoneConfig = zone

		switch id {
		case path[1]:
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_DATABASE
		case path[2]:
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_TABLE
		default:
			resp.ZoneConfigLevel = serverpb.ZoneConfigurationLevel_CLUSTER
		}
	}

	// Get the number of ranges in the table. We get the key span for the table
	// data. Then, we count the number of ranges that make up that key span.
	{
		tableSpan := generateTableSpan(tableID)
		tableRSpan := roachpb.RSpan{}
		var err error
		tableRSpan.Key, err = keys.Addr(tableSpan.Key)
		if err != nil {
			return nil, s.serverError(err)
		}
		tableRSpan.EndKey, err = keys.Addr(tableSpan.EndKey)
		if err != nil {
			return nil, s.serverError(err)
		}
		rangeCount, err := s.server.distSender.CountRanges(ctx, tableRSpan)
		if err != nil {
			return nil, s.serverError(err)
		}
		resp.RangeCount = rangeCount
	}

	return &resp, nil
}

// generateTableSpan generates a table's key span.
//
// NOTE: this doesn't make sense for interleaved (children) table. As of
// 03/2018, callers around here use it anyway.
func generateTableSpan(tableID sqlbase.ID) roachpb.Span {
	tablePrefix := keys.MakeTablePrefix(uint32(tableID))
	tableStartKey := roachpb.Key(tablePrefix)
	tableEndKey := tableStartKey.PrefixEnd()
	return roachpb.Span{Key: tableStartKey, EndKey: tableEndKey}
}

// TableStats is an endpoint that returns disk usage and replication statistics
// for the specified table.
func (s *adminServer) TableStats(
	ctx context.Context, req *serverpb.TableStatsRequest,
) (*serverpb.TableStatsResponse, error) {
	// Get table span.
	path, err := s.queryDescriptorIDPath(
		ctx, sql.SessionArgs{User: s.getUser(req)}, []string{req.Database, req.Table},
	)
	if err != nil {
		return nil, s.serverError(err)
	}
	tableID := path[2]
	tableSpan := generateTableSpan(tableID)

	return s.tableStatsForSpan(ctx, tableSpan)
}

// NonTableStats is an endpoint that returns disk usage and replication
// statistics for the time series system.
func (s *adminServer) NonTableStats(
	ctx context.Context, req *serverpb.NonTableStatsRequest,
) (*serverpb.NonTableStatsResponse, error) {
	timeSeriesStats, err := s.tableStatsForSpan(ctx, roachpb.Span{
		Key:    keys.TimeseriesPrefix,
		EndKey: keys.TimeseriesPrefix.PrefixEnd(),
	})
	if err != nil {
		return nil, err
	}
	response := serverpb.NonTableStatsResponse{
		TimeSeriesStats: timeSeriesStats,
	}
	return &response, nil
}

func (s *adminServer) tableStatsForSpan(
	ctx context.Context, tableSpan roachpb.Span,
) (*serverpb.TableStatsResponse, error) {
	startKey, err := keys.Addr(tableSpan.Key)
	if err != nil {
		return nil, s.serverError(err)
	}
	endKey, err := keys.Addr(tableSpan.EndKey)
	if err != nil {
		return nil, s.serverError(err)
	}

	// Get current range descriptors for table. This is done by scanning over
	// meta2 keys for the range.
	rangeDescKVs, err := s.server.db.Scan(ctx, keys.RangeMetaKey(startKey), keys.RangeMetaKey(endKey), 0)
	if err != nil {
		return nil, s.serverError(err)
	}

	// Extract a list of node IDs from the response.
	nodeIDs := make(map[roachpb.NodeID]struct{})
	for _, kv := range rangeDescKVs {
		var rng roachpb.RangeDescriptor
		if err := kv.Value.GetProto(&rng); err != nil {
			return nil, s.serverError(err)
		}
		for _, repl := range rng.Replicas {
			nodeIDs[repl.NodeID] = struct{}{}
		}
	}

	// Construct TableStatsResponse by sending an RPC to every node involved.
	tableStatResponse := serverpb.TableStatsResponse{
		NodeCount: int64(len(nodeIDs)),
		// TODO(mrtracy): The "RangeCount" returned by TableStats is more
		// accurate than the "RangeCount" returned by TableDetails, because this
		// method always consistently queries the meta2 key range for the table;
		// in contrast, TableDetails uses a method on the DistSender, which
		// queries using a range metadata cache and thus may return stale data
		// for tables that are rapidly splitting. However, one potential
		// *advantage* of using the DistSender is that it will populate the
		// DistSender's range metadata cache in the case where meta2 information
		// for this table is not already present; the query used by TableStats
		// does not populate the DistSender cache. We should consider plumbing
		// TableStats' meta2 query through the DistSender so that it will share
		// the advantage of populating the cache (without the disadvantage of
		// potentially returning stale data).
		// See Github #5435 for some discussion.
		RangeCount: int64(len(rangeDescKVs)),
	}
	type nodeResponse struct {
		nodeID roachpb.NodeID
		resp   *serverpb.SpanStatsResponse
		err    error
	}

	// Send a SpanStats query to each node.
	responses := make(chan nodeResponse, len(nodeIDs))
	for nodeID := range nodeIDs {
		nodeID := nodeID // avoid data race
		if err := s.server.stopper.RunAsyncTask(
			ctx, "server.adminServer: requesting remote stats",
			func(ctx context.Context) {
				// Set a generous timeout on the context for each individual query.
				ctx, cancel := context.WithTimeout(ctx, 5*base.NetworkTimeout)
				defer cancel()

				var spanResponse *serverpb.SpanStatsResponse
				client, err := s.server.status.dialNode(ctx, nodeID)
				if err == nil {
					req := serverpb.SpanStatsRequest{
						StartKey: startKey,
						EndKey:   endKey,
						NodeID:   nodeID.String(),
					}
					spanResponse, err = client.SpanStats(ctx, &req)
				}

				// Channel is buffered, can always write.
				responses <- nodeResponse{
					nodeID: nodeID,
					resp:   spanResponse,
					err:    err,
				}
			}); err != nil {
			return nil, err
		}
	}
	for remainingResponses := len(nodeIDs); remainingResponses > 0; remainingResponses-- {
		select {
		case resp := <-responses:
			// For nodes which returned an error, note that the node's data
			// is missing. For successful calls, aggregate statistics.
			if resp.err != nil {
				tableStatResponse.MissingNodes = append(
					tableStatResponse.MissingNodes,
					serverpb.TableStatsResponse_MissingNode{
						NodeID:       resp.nodeID.String(),
						ErrorMessage: resp.err.Error(),
					},
				)
			} else {
				tableStatResponse.Stats.Add(resp.resp.TotalStats)
				tableStatResponse.ReplicaCount += int64(resp.resp.RangeCount)
				tableStatResponse.ApproximateDiskBytes += resp.resp.ApproximateDiskBytes
			}
		case <-ctx.Done():
			// Caller gave up, stop doing work.
			return nil, ctx.Err()
		}
	}

	return &tableStatResponse, nil
}

// Users returns a list of users, stripped of any passwords.
func (s *adminServer) Users(
	ctx context.Context, req *serverpb.UsersRequest,
) (*serverpb.UsersResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}
	query := `SELECT username FROM system.users WHERE "isRole" = false`
	rows, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-users", nil /* txn */, args, query,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	var resp serverpb.UsersResponse
	for _, row := range rows {
		resp.Users = append(resp.Users, serverpb.UsersResponse_User{Username: string(tree.MustBeDString(row[0]))})
	}
	return &resp, nil
}

// Events is an endpoint that returns the latest event log entries, with the following
// optional URL parameters:
//
// type=STRING  returns events with this type (e.g. "create_table")
// targetID=INT returns events for that have this targetID
func (s *adminServer) Events(
	ctx context.Context, req *serverpb.EventsRequest,
) (*serverpb.EventsResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}

	limit := req.Limit
	if limit == 0 {
		limit = defaultAPIEventLimit
	}

	// Execute the query.
	q := makeSQLQuery()
	q.Append(`SELECT timestamp, "eventType", "targetID", "reportingID", info, "uniqueID" `)
	q.Append("FROM system.eventlog ")
	q.Append("WHERE true ") // This simplifies the WHERE clause logic below.
	if len(req.Type) > 0 {
		q.Append(`AND "eventType" = $ `, req.Type)
	}
	if req.TargetId > 0 {
		q.Append(`AND "targetID" = $ `, req.TargetId)
	}
	q.Append("ORDER BY timestamp DESC ")
	if limit > 0 {
		q.Append("LIMIT $", limit)
	}
	if len(q.Errors()) > 0 {
		return nil, s.serverErrors(q.Errors())
	}
	rows, cols, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-events", nil /* txn */, args, q.String(), q.QueryArguments()...)
	if err != nil {
		return nil, s.serverError(err)
	}

	// Marshal response.
	var resp serverpb.EventsResponse
	scanner := makeResultScanner(cols)
	for _, row := range rows {
		var event serverpb.EventsResponse_Event
		var ts time.Time
		if err := scanner.ScanIndex(row, 0, &ts); err != nil {
			return nil, err
		}
		event.Timestamp = ts
		if err := scanner.ScanIndex(row, 1, &event.EventType); err != nil {
			return nil, err
		}
		if err := scanner.ScanIndex(row, 2, &event.TargetID); err != nil {
			return nil, err
		}
		if err := scanner.ScanIndex(row, 3, &event.ReportingID); err != nil {
			return nil, err
		}
		if err := scanner.ScanIndex(row, 4, &event.Info); err != nil {
			return nil, err
		}
		if event.EventType == string(sql.EventLogSetClusterSetting) {

			// TODO: `if s.getUser(req) != security.RootUser` when we have auth.

			event.Info = redactSettingsChange(event.Info)
		}
		if err := scanner.ScanIndex(row, 5, &event.UniqueID); err != nil {
			return nil, err
		}

		resp.Events = append(resp.Events, event)
	}
	return &resp, nil
}

// make a best-effort attempt at redacting the setting value.
func redactSettingsChange(info string) string {
	var s sql.EventLogSetClusterSettingDetail
	if err := json.Unmarshal([]byte(info), &s); err != nil {
		return ""
	}
	s.Value = "<hidden>"
	ret, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	return string(ret)
}

// RangeLog is an endpoint that returns the latest range log entries.
func (s *adminServer) RangeLog(
	ctx context.Context, req *serverpb.RangeLogRequest,
) (*serverpb.RangeLogResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)

	args := sql.SessionArgs{User: s.getUser(req)}

	limit := req.Limit
	if limit == 0 {
		limit = defaultAPIEventLimit
	}

	includeRawKeys := debug.GatewayRemoteAllowed(ctx, s.server.ClusterSettings())

	// Execute the query.
	q := makeSQLQuery()
	q.Append(`SELECT timestamp, "rangeID", "storeID", "eventType", "otherRangeID", info `)
	q.Append("FROM system.rangelog ")
	if req.RangeId > 0 {
		rangeID := tree.NewDInt(tree.DInt(req.RangeId))
		q.Append(`WHERE "rangeID" = $ OR "otherRangeID" = $`, rangeID, rangeID)
	}
	if limit > 0 {
		q.Append("ORDER BY timestamp desc ")
		q.Append("LIMIT $", tree.NewDInt(tree.DInt(limit)))
	}
	if len(q.Errors()) > 0 {
		return nil, s.serverErrors(q.Errors())
	}
	rows, cols, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-range-log",
		nil /* txn */, args, q.String(), q.QueryArguments()...,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	// Marshal response.
	var resp serverpb.RangeLogResponse
	if len(cols) != 6 {
		return nil, errors.Errorf("incorrect number of columns in response, expected 6, got %d", len(cols))
	}
	scanner := makeResultScanner(cols)
	for _, row := range rows {
		var event storage.RangeLogEvent
		var ts time.Time
		if err := scanner.ScanIndex(row, 0, &ts); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Timestamp didn't parse correctly: %s", row[0].String()))
		}
		event.Timestamp = ts
		var rangeID int64
		if err := scanner.ScanIndex(row, 1, &rangeID); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("RangeID didn't parse correctly: %s", row[1].String()))
		}
		event.RangeID = roachpb.RangeID(rangeID)
		var storeID int64
		if err := scanner.ScanIndex(row, 2, &storeID); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("StoreID didn't parse correctly: %s", row[2].String()))
		}
		event.StoreID = roachpb.StoreID(int32(storeID))
		var eventTypeString string
		if err := scanner.ScanIndex(row, 3, &eventTypeString); err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("EventType didn't parse correctly: %s", row[3].String()))
		}
		if eventType, ok := storage.RangeLogEventType_value[eventTypeString]; ok {
			event.EventType = storage.RangeLogEventType(eventType)
		} else {
			return nil, errors.Errorf("EventType didn't parse correctly: %s", eventTypeString)
		}

		var otherRangeID int64
		if row[4].String() != "NULL" {
			if err := scanner.ScanIndex(row, 4, &otherRangeID); err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("OtherRangeID didn't parse correctly: %s", row[4].String()))
			}
			event.OtherRangeID = roachpb.RangeID(otherRangeID)
		}

		var prettyInfo serverpb.RangeLogResponse_PrettyInfo
		if row[5].String() != "NULL" {
			var info string
			if err := scanner.ScanIndex(row, 5, &info); err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("info didn't parse correctly: %s", row[5].String()))
			}
			if err := json.Unmarshal([]byte(info), &event.Info); err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("info didn't parse correctly: %s", info))
			}
			if event.Info.NewDesc != nil {
				if !includeRawKeys {
					event.Info.NewDesc.StartKey = nil
					event.Info.NewDesc.EndKey = nil
				}
				prettyInfo.NewDesc = event.Info.NewDesc.String()
			}
			if event.Info.UpdatedDesc != nil {
				if !includeRawKeys {
					event.Info.UpdatedDesc.StartKey = nil
					event.Info.UpdatedDesc.EndKey = nil
				}
				prettyInfo.UpdatedDesc = event.Info.UpdatedDesc.String()
			}
			if event.Info.AddedReplica != nil {
				prettyInfo.AddedReplica = event.Info.AddedReplica.String()
			}
			if event.Info.RemovedReplica != nil {
				prettyInfo.RemovedReplica = event.Info.RemovedReplica.String()
			}
			prettyInfo.Reason = string(event.Info.Reason)
			prettyInfo.Details = event.Info.Details
		}

		resp.Events = append(resp.Events, serverpb.RangeLogResponse_Event{
			Event:      event,
			PrettyInfo: prettyInfo,
		})
	}
	return &resp, nil
}

// getUIData returns the values and timestamps for the given UI keys. Keys
// that are not found will not be returned.
func (s *adminServer) getUIData(
	ctx context.Context, args sql.SessionArgs, user string, keys []string,
) (*serverpb.GetUIDataResponse, error) {
	if len(keys) == 0 {
		return &serverpb.GetUIDataResponse{}, nil
	}

	// Query database.
	query := makeSQLQuery()
	query.Append(`SELECT key, value, "lastUpdated" FROM system.ui WHERE key IN (`)
	for i, key := range keys {
		if i != 0 {
			query.Append(",")
		}
		query.Append("$", tree.NewDString(key))
	}
	query.Append(");")
	if err := query.Errors(); err != nil {
		return nil, s.serverErrorf("error constructing query: %v", err)
	}
	rows, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-getUIData", nil /* txn */, args, query.String(), query.QueryArguments()...,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	// Marshal results.
	resp := serverpb.GetUIDataResponse{KeyValues: make(map[string]serverpb.GetUIDataResponse_Value)}
	for _, row := range rows {
		dKey, ok := tree.AsDString(row[0])
		if !ok {
			return nil, s.serverErrorf("unexpected type for UI key: %T", row[0])
		}
		dValue, ok := row[1].(*tree.DBytes)
		if !ok {
			return nil, s.serverErrorf("unexpected type for UI value: %T", row[1])
		}
		dLastUpdated, ok := row[2].(*tree.DTimestamp)
		if !ok {
			return nil, s.serverErrorf("unexpected type for UI lastUpdated: %T", row[2])
		}

		resp.KeyValues[string(dKey)] = serverpb.GetUIDataResponse_Value{
			Value:       []byte(*dValue),
			LastUpdated: dLastUpdated.Time,
		}
	}
	return &resp, nil
}

// SetUIData is an endpoint that stores the given key/value pairs in the
// system.ui table. See GetUIData for more details on semantics.
func (s *adminServer) SetUIData(
	ctx context.Context, req *serverpb.SetUIDataRequest,
) (*serverpb.SetUIDataResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)

	if len(req.KeyValues) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "KeyValues cannot be empty")
	}

	args := sql.SessionArgs{User: s.getUser(req)}

	for key, val := range req.KeyValues {
		// Do an upsert of the key. We update each key in a separate transaction to
		// avoid long-running transactions and possible deadlocks.
		query := `UPSERT INTO system.ui (key, value, "lastUpdated") VALUES ($1, $2, now())`
		rowsAffected, err := s.server.internalExecutor.ExecWithSessionArgs(
			ctx, "admin-set-ui-data", nil /* txn */, args, query, key, val)
		if err != nil {
			return nil, s.serverError(err)
		}
		if rowsAffected != 1 {
			return nil, s.serverErrorf("rows affected %d != expected %d", rowsAffected, 1)
		}
	}

	return &serverpb.SetUIDataResponse{}, nil
}

// GetUIData returns data associated with the given keys, which was stored
// earlier through SetUIData.
//
// The stored values are meant to be opaque to the server. In the rare case that
// the server code needs to call this method, it should only read from keys that
// have the prefix `serverUIDataKeyPrefix`.
func (s *adminServer) GetUIData(
	ctx context.Context, req *serverpb.GetUIDataRequest,
) (*serverpb.GetUIDataResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}

	if len(req.Keys) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "keys cannot be empty")
	}

	resp, err := s.getUIData(ctx, args, s.getUser(req), req.Keys)
	if err != nil {
		return nil, s.serverError(err)
	}

	return resp, nil
}

// Settings returns settings associated with the given keys.
func (s *adminServer) Settings(
	ctx context.Context, req *serverpb.SettingsRequest,
) (*serverpb.SettingsResponse, error) {
	keys := req.Keys
	if len(keys) == 0 {
		keys = settings.Keys()
	}

	resp := serverpb.SettingsResponse{KeyValues: make(map[string]serverpb.SettingsResponse_Value)}
	for _, k := range keys {
		v, ok := settings.Lookup(k)
		if !ok {
			continue
		}
		resp.KeyValues[k] = serverpb.SettingsResponse_Value{
			Type:        v.Typ(),
			Value:       settings.SanitizedValue(k, &s.server.st.SV),
			Description: v.Description(),
		}
	}

	return &resp, nil
}

// Cluster returns cluster metadata.
func (s *adminServer) Cluster(
	_ context.Context, req *serverpb.ClusterRequest,
) (*serverpb.ClusterResponse, error) {
	clusterID := s.server.ClusterID()
	if clusterID == (uuid.UUID{}) {
		return nil, status.Errorf(codes.Unavailable, "cluster ID not yet available")
	}

	// Check if enterprise features are enabled.  We currently test for the
	// feature "BACKUP", although enterprise licenses do not yet distinguish
	// between different features.
	organization := sql.ClusterOrganization.Get(&s.server.st.SV)
	enterpriseEnabled := base.CheckEnterpriseEnabled(s.server.st, clusterID, organization, "BACKUP") == nil

	return &serverpb.ClusterResponse{
		ClusterID:         clusterID.String(),
		ReportingEnabled:  log.DiagnosticsReportingEnabled.Get(&s.server.st.SV),
		EnterpriseEnabled: enterpriseEnabled,
	}, nil
}

// Health returns liveness for the node target of the request.
func (s *adminServer) Health(
	ctx context.Context, req *serverpb.HealthRequest,
) (*serverpb.HealthResponse, error) {
	isLive, err := s.server.nodeLiveness.IsLive(s.server.NodeID())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !isLive {
		return nil, status.Errorf(codes.Unavailable, "node is not live")
	}
	return &serverpb.HealthResponse{}, nil
}

// Liveness returns the liveness state of all nodes on the cluster.
func (s *adminServer) Liveness(
	context.Context, *serverpb.LivenessRequest,
) (*serverpb.LivenessResponse, error) {
	livenesses := s.server.nodeLiveness.GetLivenesses()
	statusMap := s.server.nodeLiveness.GetLivenessStatusMap()

	return &serverpb.LivenessResponse{
		Livenesses: livenesses,
		Statuses:   statusMap,
	}, nil
}

func (s *adminServer) Jobs(
	ctx context.Context, req *serverpb.JobsRequest,
) (*serverpb.JobsResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}

	q := makeSQLQuery()
	q.Append(`
      SELECT id, type, description, username, descriptor_ids, status,
             created, started, finished, modified, fraction_completed, error
        FROM crdb_internal.jobs
       WHERE true
	`)
	if req.Status != "" {
		q.Append(" AND status = $", req.Status)
	}
	if req.Type != jobspb.TypeUnspecified {
		q.Append(" AND type = $", req.Type.String())
	}
	q.Append("ORDER BY created DESC")
	if req.Limit > 0 {
		q.Append(" LIMIT $", tree.DInt(req.Limit))
	}
	rows, cols, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-jobs", nil /* txn */, args, q.String(), q.QueryArguments()...,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	scanner := makeResultScanner(cols)
	resp := serverpb.JobsResponse{
		Jobs: make([]serverpb.JobsResponse_Job, len(rows)),
	}
	for i, row := range rows {
		job := &resp.Jobs[i]
		if err := scanner.ScanAll(
			row,
			&job.ID,
			&job.Type,
			&job.Description,
			&job.Username,
			&job.DescriptorIDs,
			&job.Status,
			&job.Created,
			&job.Started,
			&job.Finished,
			&job.Modified,
			&job.FractionCompleted,
			&job.Error,
		); err != nil {
			return nil, s.serverError(err)
		}
	}

	return &resp, nil
}

func (s *adminServer) Locations(
	ctx context.Context, req *serverpb.LocationsRequest,
) (*serverpb.LocationsResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}

	q := makeSQLQuery()
	q.Append(`SELECT "localityKey", "localityValue", latitude, longitude FROM system.locations`)
	rows, cols, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-locations", nil /* txn */, args, q.String(),
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	scanner := makeResultScanner(cols)
	resp := serverpb.LocationsResponse{
		Locations: make([]serverpb.LocationsResponse_Location, len(rows)),
	}
	for i, row := range rows {
		loc := &resp.Locations[i]
		lat, lon := new(apd.Decimal), new(apd.Decimal)
		if err := scanner.ScanAll(
			row, &loc.LocalityKey, &loc.LocalityValue, lat, lon); err != nil {
			return nil, s.serverError(err)
		}
		if loc.Latitude, err = lat.Float64(); err != nil {
			return nil, s.serverError(err)
		}
		if loc.Longitude, err = lon.Float64(); err != nil {
			return nil, s.serverError(err)
		}
	}

	return &resp, nil
}

// QueryPlan returns a JSON representation of a distsql physical query
// plan.
func (s *adminServer) QueryPlan(
	ctx context.Context, req *serverpb.QueryPlanRequest,
) (*serverpb.QueryPlanResponse, error) {
	ctx = s.server.AnnotateCtx(ctx)
	args := sql.SessionArgs{User: s.getUser(req)}

	// As long as there's only one query provided it's safe to construct the
	// explain query.
	stmts, err := parser.Parse(req.Query)
	if err != nil {
		return nil, s.serverError(err)
	}
	if len(stmts) > 1 {
		return nil, s.serverErrorf("more than one query provided")
	}

	explain := fmt.Sprintf(
		"SELECT \"JSON\" FROM [EXPLAIN (distsql) %s]",
		strings.Trim(req.Query, ";"))
	rows, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-query-plan", nil /* txn */, args, explain,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	row := rows[0]
	dbDatum, ok := tree.AsDString(row[0])
	if !ok {
		return nil, s.serverErrorf("type assertion failed on json: %T", row)
	}

	return &serverpb.QueryPlanResponse{
		DistSQLPhysicalQueryPlan: string(dbDatum),
	}, nil
}

// Drain puts the node into the specified drain mode(s) and optionally
// instructs the process to terminate.
func (s *adminServer) Drain(req *serverpb.DrainRequest, stream serverpb.Admin_DrainServer) error {
	on := make([]serverpb.DrainMode, len(req.On))
	for i := range req.On {
		on[i] = serverpb.DrainMode(req.On[i])
	}
	off := make([]serverpb.DrainMode, len(req.Off))
	for i := range req.Off {
		off[i] = serverpb.DrainMode(req.Off[i])
	}

	ctx := stream.Context()
	_ = s.server.Undrain(ctx, off)

	nowOn, err := s.server.Drain(ctx, on)
	if err != nil {
		return err
	}

	res := serverpb.DrainResponse{
		On: make([]int32, len(nowOn)),
	}
	for i := range nowOn {
		res.On[i] = int32(nowOn[i])
	}
	if err := stream.Send(&res); err != nil {
		return err
	}

	if !req.Shutdown {
		return nil
	}

	s.server.grpc.Stop()

	go func() {
		// The explicit closure here allows callers.Lookup() to return something
		// sensible referring to this file (otherwise it ends up in runtime
		// internals).
		s.server.stopper.Stop(ctx)
	}()

	select {
	case <-s.server.stopper.IsStopped():
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// DecommissionStatus returns the DecommissionStatus for all or the given nodes.
func (s *adminServer) DecommissionStatus(
	ctx context.Context, req *serverpb.DecommissionStatusRequest,
) (*serverpb.DecommissionStatusResponse, error) {
	// Get the number of replicas on each node. We *may* not need all of them,
	// but that would be more complicated than seems worth it right now.
	ns, err := s.server.status.Nodes(ctx, &serverpb.NodesRequest{})
	if err != nil {
		return nil, errors.Wrap(err, "loading node statuses")
	}

	nodeIDs := req.NodeIDs
	// If no nodeIDs given, use all nodes.
	if len(nodeIDs) == 0 {
		for _, status := range ns.Nodes {
			nodeIDs = append(nodeIDs, status.Desc.NodeID)
		}
	}

	// Compute the replica counts for the target nodes only. This map doubles as
	// a lookup table to check whether we care about a given node.
	replicaCounts := make(map[roachpb.NodeID]int64)
	for _, nodeID := range nodeIDs {
		replicaCounts[nodeID] = math.MaxInt64
	}

	for _, nodeStatus := range ns.Nodes {
		nodeID := nodeStatus.Desc.NodeID
		if _, ok := replicaCounts[nodeID]; !ok {
			continue // not interested in this node
		}
		var replicas float64
		for _, storeStatus := range nodeStatus.StoreStatuses {
			replicas += storeStatus.Metrics["replicas"]
		}
		replicaCounts[nodeID] = int64(replicas)
	}

	var res serverpb.DecommissionStatusResponse

	for nodeID := range replicaCounts {
		l, err := s.server.nodeLiveness.GetLiveness(nodeID)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to get liveness for %d", nodeID)
		}
		nodeResp := serverpb.DecommissionStatusResponse_Status{
			NodeID:          l.NodeID,
			ReplicaCount:    replicaCounts[l.NodeID],
			Decommissioning: l.Decommissioning,
			Draining:        l.Draining,
		}
		if l.IsLive(s.server.clock.Now(), s.server.clock.MaxOffset()) {
			nodeResp.IsLive = true
		}

		res.Status = append(res.Status, nodeResp)
	}

	sort.Slice(res.Status, func(i, j int) bool {
		return res.Status[i].NodeID < res.Status[j].NodeID
	})

	return &res, nil
}

// Decommission sets the decommission flag to the specified value on the specified node(s).
func (s *adminServer) Decommission(
	ctx context.Context, req *serverpb.DecommissionRequest,
) (*serverpb.DecommissionStatusResponse, error) {
	nodeIDs := req.NodeIDs
	if nodeIDs == nil {
		// If no NodeIDs are specified, decommission the current node. This is
		// used by `quit --decommission`.
		nodeIDs = []roachpb.NodeID{s.server.NodeID()}
	}

	// Mark the target nodes as decommissioning. They'll find out as they
	// heartbeat their liveness.
	if err := s.server.Decommission(ctx, req.Decommissioning, nodeIDs); err != nil {
		return nil, err
	}
	return s.DecommissionStatus(ctx, &serverpb.DecommissionStatusRequest{NodeIDs: nodeIDs})
}

// DataDistribution returns a count of replicas on each node for each table.
func (s *adminServer) DataDistribution(
	ctx context.Context, req *serverpb.DataDistributionRequest,
) (*serverpb.DataDistributionResponse, error) {
	resp := &serverpb.DataDistributionResponse{
		DatabaseInfo: make(map[string]serverpb.DataDistributionResponse_DatabaseInfo),
		ZoneConfigs:  make(map[int64]serverpb.DataDistributionResponse_ZoneConfig),
	}

	// Get ids and names for databases and tables.
	// Set up this structure in the response.

	// This relies on crdb_internal.tables returning data even for newly added tables
	// and deleted tables (as opposed to e.g. information_schema) because we are interested
	// in the data for all ranges, not just ranges for visible tables.
	args := sql.SessionArgs{User: s.getUser(req)}
	tablesQuery := `
		SELECT name, table_id, database_name, parent_id
		FROM "".crdb_internal.tables
		WHERE state = 'PUBLIC'
	`
	rows1, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-replica-matrix", nil /* txn */, args, tablesQuery,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	// Used later when we're scanning Meta2 and only have IDs, not names.
	tableInfosByTableID := map[uint64]serverpb.DataDistributionResponse_TableInfo{}

	for _, row := range rows1 {
		tableName := (*string)(row[0].(*tree.DString))
		tableID := uint64(tree.MustBeDInt(row[1]))
		dbName := (*string)(row[2].(*tree.DString))

		// Insert database if it doesn't exist.
		dbInfo, ok := resp.DatabaseInfo[*dbName]
		if !ok {
			dbInfo = serverpb.DataDistributionResponse_DatabaseInfo{
				TableInfo: make(map[string]serverpb.DataDistributionResponse_TableInfo),
			}
			resp.DatabaseInfo[*dbName] = dbInfo
		}

		// Get zone config for table.
		zoneConfigQuery := fmt.Sprintf(
			`SELECT id, cli_specifier FROM [EXPERIMENTAL SHOW ZONE CONFIGURATION FOR TABLE %s.%s]`,
			(*tree.Name)(dbName), (*tree.Name)(tableName),
		)
		rows, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
			ctx, "admin-replica-matrix", nil /* txn */, args, zoneConfigQuery,
		)
		if err != nil {
			return nil, s.serverError(err)
		}

		if len(rows) != 1 {
			return nil, s.serverError(fmt.Errorf(
				"could not get zone config for table %s; %d rows returned", *tableName, len(rows),
			))
		}

		zcRow := rows[0]
		zcID := int64(tree.MustBeDInt(zcRow[0]))

		// Insert table.
		tableInfo := serverpb.DataDistributionResponse_TableInfo{
			ReplicaCountByNodeId: make(map[roachpb.NodeID]int64),
			ZoneConfigId:         zcID,
		}
		dbInfo.TableInfo[*tableName] = tableInfo
		tableInfosByTableID[tableID] = tableInfo
	}

	// Get replica counts.
	if err := s.server.db.Txn(ctx, func(txnCtx context.Context, txn *client.Txn) error {
		acct := s.memMonitor.MakeBoundAccount()
		defer acct.Close(txnCtx)

		kvs, err := sql.ScanMetaKVs(ctx, txn, roachpb.Span{
			Key:    keys.UserTableDataMin,
			EndKey: keys.MaxKey,
		})
		if err != nil {
			return err
		}

		// Group replicas by table and node, accumulate counts.
		var rangeDesc roachpb.RangeDescriptor
		for _, kv := range kvs {
			if err := acct.Grow(txnCtx, int64(len(kv.Key)+len(kv.Value.RawBytes))); err != nil {
				return err
			}
			if err := kv.ValueProto(&rangeDesc); err != nil {
				return err
			}

			_, tableID, err := keys.DecodeTablePrefix(rangeDesc.StartKey.AsRawKey())
			if err != nil {
				return err
			}

			for _, replicaDesc := range rangeDesc.Replicas {
				tableInfo, ok := tableInfosByTableID[tableID]
				if !ok {
					// This is a database, skip.
					continue
				}
				tableInfo.ReplicaCountByNodeId[replicaDesc.NodeID]++
			}
		}
		return nil
	}); err != nil {
		return nil, s.serverError(err)
	}

	// Get zone configs.
	// TODO(vilterp): this can be done in parallel with getting table/db names and replica counts.
	zoneConfigsQuery := `EXPERIMENTAL SHOW ALL ZONE CONFIGURATIONS`
	rows2, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-replica-matrix", nil /* txn */, args, zoneConfigsQuery,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	for _, row := range rows2 {
		zcID := int64(tree.MustBeDInt(row[0]))
		zcCliSpecifier := string(tree.MustBeDString(row[1]))
		zcYaml := tree.MustBeDBytes(row[2])
		zcBytes := tree.MustBeDBytes(row[3])
		var zcProto config.ZoneConfig
		if err := protoutil.Unmarshal([]byte(zcBytes), &zcProto); err != nil {
			return nil, s.serverError(err)
		}

		resp.ZoneConfigs[zcID] = serverpb.DataDistributionResponse_ZoneConfig{
			CliSpecifier: zcCliSpecifier,
			Config:       zcProto,
			ConfigYaml:   string(zcYaml),
		}
	}

	return resp, nil
}

func (s *adminServer) Queries(
	ctx context.Context, req *serverpb.QueriesRequest,
) (*serverpb.QueriesResponse, error) {
	stmtStats := s.server.pgServer.SQLServer.GetUnscrubbedStmtStats()
	lastReset := s.server.pgServer.SQLServer.GetStmtStatsLastReset()

	return &serverpb.QueriesResponse{
		Queries:   stmtStats,
		LastReset: lastReset,
	}, nil
}

// sqlQuery allows you to incrementally build a SQL query that uses
// placeholders. Instead of specific placeholders like $1, you instead use the
// temporary placeholder $.
type sqlQuery struct {
	buf   bytes.Buffer
	pidx  int
	qargs []interface{}
	errs  []error
}

func makeSQLQuery() *sqlQuery {
	res := &sqlQuery{}
	return res
}

// String returns the full query.
func (q *sqlQuery) String() string {
	if len(q.errs) > 0 {
		return "couldn't generate query: please check Errors()"
	}
	return q.buf.String()
}

// Errors returns a slice containing all errors that have happened during the
// construction of this query.
func (q *sqlQuery) Errors() []error {
	return q.errs
}

// QueryArguments returns a filled map of placeholders containing all arguments
// provided to this query through Append.
func (q *sqlQuery) QueryArguments() []interface{} {
	return q.qargs
}

// Append appends the provided string and any number of query parameters.
// Instead of using normal placeholders (e.g. $1, $2), use meta-placeholder $.
// This method rewrites the query so that it uses proper placeholders.
//
// For example, suppose we have the following calls:
//
//   query.Append("SELECT * FROM foo WHERE a > $ AND a < $ ", arg1, arg2)
//   query.Append("LIMIT $", limit)
//
// The query is rewritten into:
//
//   SELECT * FROM foo WHERE a > $1 AND a < $2 LIMIT $3
//   /* $1 = arg1, $2 = arg2, $3 = limit */
//
// Note that this method does NOT return any errors. Instead, we queue up
// errors, which can later be accessed. Returning an error here would make
// query construction code exceedingly tedious.
func (q *sqlQuery) Append(s string, params ...interface{}) {
	var placeholders int
	for _, r := range s {
		q.buf.WriteRune(r)
		if r == '$' {
			q.pidx++
			placeholders++
			q.buf.WriteString(strconv.Itoa(q.pidx)) // SQL placeholders are 1-based
		}
	}

	if placeholders != len(params) {
		q.errs = append(q.errs,
			errors.Errorf("# of placeholders %d != # of params %d", placeholders, len(params)))
	}
	q.qargs = append(q.qargs, params...)
}

// resultScanner scans columns from sql.ResultRow instances into variables,
// performing the appropriate casting and error detection along the way.
type resultScanner struct {
	colNameToIdx map[string]int
}

func makeResultScanner(cols []sqlbase.ResultColumn) resultScanner {
	rs := resultScanner{
		colNameToIdx: make(map[string]int),
	}
	for i, col := range cols {
		rs.colNameToIdx[col.Name] = i
	}
	return rs
}

// IsNull returns whether the specified column of the given row contains
// a SQL NULL value.
func (rs resultScanner) IsNull(row tree.Datums, col string) (bool, error) {
	idx, ok := rs.colNameToIdx[col]
	if !ok {
		return false, errors.Errorf("result is missing column %s", col)
	}
	return row[idx] == tree.DNull, nil
}

// ScanIndex scans the given column index of the given row into dst.
func (rs resultScanner) ScanIndex(row tree.Datums, index int, dst interface{}) error {
	src := row[index]

	if dst == nil {
		return errors.Errorf("nil destination pointer passed in")
	}

	switch d := dst.(type) {
	case *string:
		s, ok := tree.AsDString(src)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		*d = string(s)

	case *bool:
		s, ok := src.(*tree.DBool)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		*d = bool(*s)

	case *float32:
		s, ok := src.(*tree.DFloat)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		*d = float32(*s)

	case *int64:
		s, ok := tree.AsDInt(src)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		*d = int64(s)

	case *[]sqlbase.ID:
		s, ok := tree.AsDArray(src)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		for i := 0; i < s.Len(); i++ {
			id, ok := tree.AsDInt(s.Array[i])
			if !ok {
				return errors.Errorf("source type assertion failed on index %d", i)
			}
			*d = append(*d, sqlbase.ID(id))
		}

	case *time.Time:
		s, ok := src.(*tree.DTimestamp)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		*d = s.Time

	// Passing a **time.Time instead of a *time.Time means the source is allowed
	// to be NULL, in which case nil is stored into *src.
	case **time.Time:
		s, ok := src.(*tree.DTimestamp)
		if !ok {
			if src != tree.DNull {
				return errors.Errorf("source type assertion failed")
			}
			*d = nil
			return nil
		}
		*d = &s.Time

	case *[]byte:
		s, ok := src.(*tree.DBytes)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		// Yes, this copies, but this probably isn't in the critical path.
		*d = []byte(*s)

	case *apd.Decimal:
		s, ok := src.(*tree.DDecimal)
		if !ok {
			return errors.Errorf("source type assertion failed")
		}
		*d = s.Decimal

	default:
		return errors.Errorf("unimplemented type for scanCol: %T", dst)
	}

	return nil
}

// ScanAll scans all the columns from the given row, in order, into dsts.
func (rs resultScanner) ScanAll(row tree.Datums, dsts ...interface{}) error {
	if len(row) != len(dsts) {
		return fmt.Errorf(
			"ScanAll: row has %d columns but %d dests provided", len(row), len(dsts))
	}
	for i := 0; i < len(row); i++ {
		if err := rs.ScanIndex(row, i, dsts[i]); err != nil {
			return err
		}
	}
	return nil
}

// Scan scans the column with the given name from the given row into dst.
func (rs resultScanner) Scan(row tree.Datums, colName string, dst interface{}) error {
	idx, ok := rs.colNameToIdx[colName]
	if !ok {
		return errors.Errorf("result is missing column %s", colName)
	}
	return rs.ScanIndex(row, idx, dst)
}

// TODO(mrtracy): The following methods, used to look up the zone configuration
// for a database or table, use the same algorithm as a set of methods in
// cli/zone.go for the same purpose. However, as that code connects to the
// server with a SQL connections, while this code uses the InternalExecutor, the
// code cannot be commonized.
//
// queryZone retrieves the specific ZoneConfig associated with the supplied ID,
// if it exists.
func (s *adminServer) queryZone(
	ctx context.Context, sargs sql.SessionArgs, id sqlbase.ID,
) (config.ZoneConfig, bool, error) {
	const query = `SELECT config FROM system.zones WHERE id = $1`
	rows, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-query-zone", nil /* txn */, sargs, query, id,
	)
	if err != nil {
		return config.ZoneConfig{}, false, err
	}

	if len(rows) == 0 {
		return config.ZoneConfig{}, false, nil
	}

	var zoneBytes []byte
	scanner := resultScanner{}
	err = scanner.ScanIndex(rows[0], 0, &zoneBytes)
	if err != nil {
		return config.ZoneConfig{}, false, err
	}

	var zone config.ZoneConfig
	if err := protoutil.Unmarshal(zoneBytes, &zone); err != nil {
		return config.ZoneConfig{}, false, err
	}
	return zone, true, nil
}

// queryZonePath queries a path of sql object IDs, as generated by
// queryDescriptorIDPath(), for a ZoneConfig. It returns the most specific
// ZoneConfig specified for the object IDs in the path.
func (s *adminServer) queryZonePath(
	ctx context.Context, sargs sql.SessionArgs, path []sqlbase.ID,
) (sqlbase.ID, config.ZoneConfig, bool, error) {
	for i := len(path) - 1; i >= 0; i-- {
		zone, zoneExists, err := s.queryZone(ctx, sargs, path[i])
		if err != nil || zoneExists {
			return path[i], zone, true, err
		}
	}
	return 0, config.ZoneConfig{}, false, nil
}

// queryNamespaceID queries for the ID of the namespace with the given name and
// parent ID.
func (s *adminServer) queryNamespaceID(
	ctx context.Context, sargs sql.SessionArgs, parentID sqlbase.ID, name string,
) (sqlbase.ID, error) {
	const query = `SELECT id FROM system.namespace WHERE "parentID" = $1 AND name = $2`
	rows, _ /* cols */, err := s.server.internalExecutor.QueryWithSessionArgs(
		ctx, "admin-query-namespace-ID", nil /* txn */, sargs, query, parentID, name,
	)
	if err != nil {
		return 0, err
	}

	if len(rows) == 0 {
		return 0, errors.Errorf("namespace %s with ParentID %d not found", name, parentID)
	}

	var id int64
	scanner := resultScanner{}
	err = scanner.ScanIndex(rows[0], 0, &id)
	if err != nil {
		return 0, err
	}

	return sqlbase.ID(id), nil
}

// queryDescriptorIDPath converts a path of namespaces into a path of namespace
// IDs. For example, if this function is called with a database/table name pair,
// it will return a list of IDs consisting of the root namespace ID, the
// databases ID, and the table ID (in that order).
func (s *adminServer) queryDescriptorIDPath(
	ctx context.Context, sargs sql.SessionArgs, names []string,
) ([]sqlbase.ID, error) {
	path := []sqlbase.ID{keys.RootNamespaceID}
	for _, name := range names {
		id, err := s.queryNamespaceID(ctx, sargs, path[len(path)-1], name)
		if err != nil {
			return nil, err
		}
		path = append(path, id)
	}
	return path, nil
}
