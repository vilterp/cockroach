package serverpb

import "github.com/cockroachdb/cockroach/pkg/roachpb"

type SchemaNode interface {
	StartKey() roachpb.Key
	EndKey() roachpb.Key
}

var _ SchemaNode = &DataDistributionResponse_DatabaseInfo{}

// hm, tables aren't contiguous within a database :(

func (di *DataDistributionResponse_DatabaseInfo) StartKey() roachpb.Key {
	return nil
}

func (di *DataDistributionResponse_DatabaseInfo) EndKey() roachpb.Key {
	return nil
}

var _ SchemaNode = &DataDistributionResponse_TableInfo{}

func (di *DataDistributionResponse_TableInfo) StartKey() roachpb.Key {
	return nil
}

func (di *DataDistributionResponse_TableInfo) EndKey() roachpb.Key {
	return nil
}

var _ SchemaNode = &DataDistributionResponse_IndexInfo{}

func (di *DataDistributionResponse_IndexInfo) StartKey() roachpb.Key {
	return nil
}

func (di *DataDistributionResponse_IndexInfo) EndKey() roachpb.Key {
	return nil
}

var _ SchemaNode = &DataDistributionResponse_PartitionInfo{}

func (di *DataDistributionResponse_PartitionInfo) StartKey() roachpb.Key {
	return nil
}

func (di *DataDistributionResponse_PartitionInfo) EndKey() roachpb.Key {
	return nil
}
