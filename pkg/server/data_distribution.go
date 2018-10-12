package server

import (
	"context"

	"fmt"

	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/internal/client"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlbase"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

func (s *adminServer) getSchemaTree(
	ctx context.Context,
) ([]*serverpb.DataDistributionResponse_DatabaseInfo, error) {
	var output []*serverpb.DataDistributionResponse_DatabaseInfo
	indexes := map[sqlbase.ID]int{} // database id => index in output

	if err := s.server.db.Txn(ctx, func(txnCtx context.Context, txn *client.Txn) error {
		descriptors, err := sql.GetAllDescriptors(ctx, txn)
		if err != nil {
			return err
		}

		// Fill in databases
		for _, descriptor := range descriptors {
			switch tDescriptor := descriptor.(type) {
			case *sqlbase.TableDescriptor:
				continue // fill these in in next pass
			case *sqlbase.DatabaseDescriptor:
				indexes[tDescriptor.ID] = len(output)
				output = append(output, &serverpb.DataDistributionResponse_DatabaseInfo{
					ID:   tDescriptor.ID,
					Name: tDescriptor.Name,
				})
			}
		}

		// Fill in tables
		for _, descriptor := range descriptors {
			switch tDescriptor := descriptor.(type) {
			case *sqlbase.TableDescriptor:
				dbInfo := output[indexes[tDescriptor.ParentID]]
				tableTree, err := s.getTableTree(ctx, *tDescriptor)
				if err != nil {
					return err
				}
				dbInfo.Tables = append(dbInfo.Tables, tableTree)
			case *sqlbase.DatabaseDescriptor:
				continue // already got these
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return output, nil
}

func (s *adminServer) getTableTree(
	ctx context.Context, desc sqlbase.TableDescriptor,
) (*serverpb.DataDistributionResponse_TableInfo, error) {
	output := &serverpb.DataDistributionResponse_TableInfo{
		ID:   desc.ID,
		Name: desc.Name,
		// TODO: dropped_at, etc
	}
	indexInfo, err := s.getIndexTree(ctx, desc, desc.PrimaryIndex)
	if err != nil {
		return nil, err
	}
	output.Indexes = append(output.Indexes, indexInfo)
	for _, index := range desc.Indexes {
		indexInfo, err := s.getIndexTree(ctx, desc, index)
		if err != nil {
			return nil, err
		}
		output.Indexes = append(output.Indexes, indexInfo)
	}
	return output, nil
}

func (s *adminServer) getIndexTree(
	ctx context.Context, tableDesc sqlbase.TableDescriptor, index sqlbase.IndexDescriptor,
) (*serverpb.DataDistributionResponse_IndexInfo, error) {
	indexSpan := tableDesc.IndexSpan(index.ID)
	fmt.Println("index span", tableDesc.ID, index.ID, indexSpan)
	indexStats, err := s.tableStatsForSpan(ctx, indexSpan)
	if err != nil {
		return nil, err
	}
	return &serverpb.DataDistributionResponse_IndexInfo{
		ID:         index.ID,
		Name:       index.Name,
		Partitions: getPartitionTree(index.Partitioning),
		Stats:      indexStats,
	}, nil
}

func getPartitionTree(
	partDesc sqlbase.PartitioningDescriptor,
) []*serverpb.DataDistributionResponse_PartitionInfo {
	var output []*serverpb.DataDistributionResponse_PartitionInfo
	if partDesc.NumColumns == 0 {
		// no partitioning. we're done.
	} else {
		if len(partDesc.Range) > 0 {
			// range partitioning
			for _, rangePartDesc := range partDesc.Range {
				output = append(output, &serverpb.DataDistributionResponse_PartitionInfo{
					Name: rangePartDesc.Name,
				})
			}
		} else {
			// list partitioning
			for _, listPartDesc := range partDesc.List {
				partInfo := &serverpb.DataDistributionResponse_PartitionInfo{
					Name: listPartDesc.Name,
				}
				partInfo.SubPartitions = getPartitionTree(listPartDesc.Subpartitioning)
				output = append(output, partInfo)
			}
		}
	}
	return output
}

// DataDistribution returns a count of replicas on each node for each table.
func (s *adminServer) DataDistribution(
	ctx context.Context, req *serverpb.DataDistributionRequest,
) (*serverpb.DataDistributionResponse, error) {
	userName := s.getUser(req)

	schemaTree, err := s.getSchemaTree(ctx)
	if err != nil {
		return nil, err
	}

	resp := &serverpb.DataDistributionResponse{
		Databases:   schemaTree,
		ZoneConfigs: map[string]serverpb.DataDistributionResponse_ZoneConfig{},
	}

	//resp := &serverpb.DataDistributionResponse{
	//	DatabaseInfo: make(map[int32]serverpb.DataDistributionResponse_DatabaseInfo),
	//	ZoneConfigs:  make(map[int64]serverpb.DataDistributionResponse_ZoneConfig),
	//}
	//
	//// Get ids and names for databases and tables.
	//// Set up this structure in the response.
	//
	//// This relies on crdb_internal.tables returning data even for newly added tables
	//// and deleted tables (as opposed to e.g. information_schema) because we are interested
	//// in the data for all ranges, not just ranges for visible tables.
	//tablesQuery := `SELECT name, table_id, database_name, drop_time FROM "".crdb_internal.tables`
	//rows1, _ /* cols */, err := s.server.internalExecutor.QueryWithUser(
	//	ctx, "admin-replica-matrix", nil /* txn */, userName, tablesQuery,
	//)
	//if err != nil {
	//	return nil, s.serverError(err)
	//}
	//
	//// Used later when we're scanning Meta2 and only have IDs, not names.
	//tableInfosByTableID := map[uint64]serverpb.DataDistributionResponse_TableInfo{}
	//
	//for _, row := range rows1 {
	//	tableName := (*string)(row[0].(*tree.DString))
	//	tableID := uint64(tree.MustBeDInt(row[1]))
	//	dbName := (*string)(row[2].(*tree.DString))
	//
	//	// Look at whether it was dropped.
	//	var droppedAtTime *time.Time
	//	droppedAtDatum, ok := row[3].(*tree.DTimestamp)
	//	if ok {
	//		droppedAtTime = &droppedAtDatum.Time
	//	}
	//
	//	// Insert database if it doesn't exist.
	//	dbInfo, ok := resp.DatabaseInfo[*dbName]
	//	if !ok {
	//		dbInfo = serverpb.DataDistributionResponse_DatabaseInfo{
	//			TableInfo: make(map[string]serverpb.DataDistributionResponse_TableInfo),
	//		}
	//		resp.DatabaseInfo[*dbName] = dbInfo
	//	}
	//
	//	// Get zone config for table.
	//	zcID := int64(0)
	//
	//	if droppedAtTime == nil {
	//		// TODO(vilterp): figure out a way to get zone configs for tables that are dropped
	//		zoneConfigQuery := fmt.Sprintf(
	//			`SELECT zone_id, cli_specifier FROM [SHOW ZONE CONFIGURATION FOR TABLE %s.%s]`,
	//			(*tree.Name)(dbName), (*tree.Name)(tableName),
	//		)
	//		rows, _ /* cols */, err := s.server.internalExecutor.QueryWithUser(
	//			ctx, "admin-replica-matrix", nil /* txn */, userName, zoneConfigQuery,
	//		)
	//		if err != nil {
	//			return nil, s.serverError(err)
	//		}
	//
	//		if len(rows) != 1 {
	//			return nil, s.serverError(fmt.Errorf(
	//				"could not get zone config for table %s; %d rows returned", *tableName, len(rows),
	//			))
	//		}
	//		zcRow := rows[0]
	//		zcID = int64(tree.MustBeDInt(zcRow[0]))
	//	}
	//
	//	// Insert table.
	//	tableInfo := serverpb.DataDistributionResponse_TableInfo{
	//		ReplicaCountByNodeId: make(map[roachpb.NodeID]int64),
	//		ZoneConfigId:         zcID,
	//		DroppedAt:            droppedAtTime,
	//	}
	//	dbInfo.TableInfo[*tableName] = tableInfo
	//	tableInfosByTableID[tableID] = tableInfo
	//}
	//
	//// Get replica counts.
	//if err := s.server.db.Txn(ctx, func(txnCtx context.Context, txn *client.Txn) error {
	//	acct := s.memMonitor.MakeBoundAccount()
	//	defer acct.Close(txnCtx)
	//
	//	kvs, err := sql.ScanMetaKVs(ctx, txn, roachpb.Span{
	//		Key:    keys.UserTableDataMin,
	//		EndKey: keys.MaxKey,
	//	})
	//	if err != nil {
	//		return err
	//	}
	//
	//	// Group replicas by table and node, accumulate counts.
	//	var rangeDesc roachpb.RangeDescriptor
	//	for _, kv := range kvs {
	//		if err := acct.Grow(txnCtx, int64(len(kv.Key)+len(kv.Value.RawBytes))); err != nil {
	//			return err
	//		}
	//		if err := kv.ValueProto(&rangeDesc); err != nil {
	//			return err
	//		}
	//
	//		_, tableID, err := keys.DecodeTablePrefix(rangeDesc.StartKey.AsRawKey())
	//		if err != nil {
	//			return err
	//		}
	//
	//		for _, replicaDesc := range rangeDesc.Replicas {
	//			tableInfo, ok := tableInfosByTableID[tableID]
	//			if !ok {
	//				// This is a database, skip.
	//				continue
	//			}
	//			tableInfo.ReplicaCountByNodeId[replicaDesc.NodeID]++
	//		}
	//	}
	//	return nil
	//}); err != nil {
	//	return nil, s.serverError(err)
	//}

	// Get zone configs.
	// TODO(vilterp): this can be done in parallel with getting table/db names and replica counts.
	zoneConfigsQuery := `
		SELECT zone_name, config_sql, config_protobuf 
		FROM crdb_internal.zones
		WHERE zone_name IS NOT NULL
	`
	rows2, _ /* cols */, err := s.server.internalExecutor.QueryWithUser(
		ctx, "admin-replica-matrix", nil /* txn */, userName, zoneConfigsQuery,
	)
	if err != nil {
		return nil, s.serverError(err)
	}

	for _, row := range rows2 {
		zoneName := string(tree.MustBeDString(row[0]))
		zcSQL := tree.MustBeDString(row[1])
		zcBytes := tree.MustBeDBytes(row[2])
		var zcProto config.ZoneConfig
		if err := protoutil.Unmarshal([]byte(zcBytes), &zcProto); err != nil {
			return nil, s.serverError(err)
		}

		resp.ZoneConfigs[zoneName] = serverpb.DataDistributionResponse_ZoneConfig{
			ZoneName:  zoneName,
			Config:    zcProto,
			ConfigSQL: string(zcSQL),
		}
	}

	return resp, nil
}
