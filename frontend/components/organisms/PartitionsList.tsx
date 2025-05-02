import { Partition } from "@/types/Partition";
import PartitionListItem from "../molecules/PartitionListItem";
import React from "react";

interface PartitionsListProps {
  partitions: Partition[];
  selectedPartition: string | null;
  onPartitionSelect: (name: string) => void;
}

const PartitionsList: React.FC<PartitionsListProps> = ({
  partitions,
  selectedPartition,
  onPartitionSelect,
}) => {
  return (
    <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden ring-1 ring-neutral-700">
      <table className="w-full">
        <thead>
          <tr className="border-b border-neutral-700">
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Nombre
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Tipo
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Tamaño
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Estado
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Montada
            </th>
          </tr>
        </thead>
        <tbody>
          {partitions.map((partition) => (
            <PartitionListItem
              key={partition.name}
              partition={partition}
              isSelected={selectedPartition === partition.name}
              onClick={onPartitionSelect}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default PartitionsList;
