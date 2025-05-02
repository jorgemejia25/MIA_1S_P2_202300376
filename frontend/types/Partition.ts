export type Partition = {
  name: string;
  type: string;
  status: string;
  fit: string;
  start: number;
  size: number;
  isMounted: boolean;
  mountId: string;
};

export type LogicalPartition = {
  name: string;
  status: string;
  fit: string;
  start: number;
  size: number;
  next: number;
  isMounted: boolean;
  mountId: string;
};

export type PartitionData = {
  partitions: Partition[];
  logicalPartitions: LogicalPartition[];
};
