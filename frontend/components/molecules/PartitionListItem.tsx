import { HiCheckCircle, HiXCircle } from "react-icons/hi";

import { Partition } from "@/types/Partition";
import React from "react";
import { formatBytes } from "@/utils/formatBytes";

interface PartitionListItemProps {
  partition: Partition;
  isSelected: boolean;
  onClick: (name: string) => void;
}

const PartitionListItem: React.FC<PartitionListItemProps> = ({
  partition,
  isSelected,
  onClick,
}) => {
  return (
    <tr
      onClick={() => onClick(partition.name)}
      className={`
        border-b border-neutral-700/50 hover:bg-neutral-700/30 cursor-pointer transition-colors
        ${isSelected ? "bg-blue-900/20" : ""}
      `}
    >
      <td className="py-3 px-4 font-medium">{partition.name}</td>
      <td className="py-3 px-4 text-sm text-gray-400">{partition.type}</td>
      <td className="py-3 px-4">
        <span className="px-2 py-1 rounded-full text-xs bg-emerald-900/30 text-emerald-300">
          {formatBytes(partition.size)}
        </span>
      </td>
      <td className="py-3 px-4 text-sm">
        <span
          className={`px-2 py-1 rounded-full text-xs 
          ${
            partition.status === "active"
              ? "bg-green-900/30 text-green-300"
              : "bg-gray-700/30 text-gray-300"
          }`}
        >
          {partition.status}
        </span>
      </td>
      <td className="py-3 px-4 text-sm">
        {partition.isMounted ? (
          <span className="flex items-center">
            <HiCheckCircle className="w-4 h-4 mr-1 text-green-400" />
            <span className="text-green-300">{partition.mountId}</span>
          </span>
        ) : (
          <span className="flex items-center">
            <HiXCircle className="w-4 h-4 mr-1 text-gray-500" />
            <span className="text-gray-400">No montada</span>
          </span>
        )}
      </td>
    </tr>
  );
};

export default PartitionListItem;
