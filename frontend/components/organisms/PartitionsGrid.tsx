import { Partition } from "@/types/Partition";
import PartitionCard from "../molecules/PartitionCard";
import React from "react";

interface PartitionsGridProps {
  partitions: Partition[];
  selectedPartition: string | null;
  onPartitionSelect: (name: string) => void;
}

const PartitionsGrid: React.FC<PartitionsGridProps> = ({
  partitions,
  selectedPartition,
  onPartitionSelect,
}) => {
  return (
    <div className="grid md:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-8">
      {partitions.map((partition) => (
        <PartitionCard
          key={partition.name}
          partition={partition}
          isSelected={selectedPartition === partition.name}
          onClick={onPartitionSelect}
        />
      ))}
    </div>
  );
};

export default PartitionsGrid;
