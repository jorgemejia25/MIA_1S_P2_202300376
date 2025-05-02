import { Disk } from "@/types/Disk";
import React from "react";
import { formatBytes } from "@/utils/formatBytes";

interface DiskListItemProps {
  disk: Disk;
  isSelected: boolean;
  onClick: (path: string) => void;
}

const DiskListItem: React.FC<DiskListItemProps> = ({
  disk,
  isSelected,
  onClick,
}) => {
  // Format date to a readable string
  const formatDate = (date: Date) => {
    return (
      date.toLocaleDateString() +
      " " +
      date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    );
  };

  return (
    <tr
      onClick={() => onClick(disk.path)}
      className={`
        border-b border-neutral-700/50 hover:bg-neutral-700/30 cursor-pointer transition-colors
        ${isSelected ? "bg-blue-900/20" : ""}
      `}
    >
      <td className="py-3 px-4 font-medium">{disk.name}</td>
      <td className="py-3 px-4 text-sm font-mono text-gray-400">{disk.path}</td>
      <td className="py-3 px-4">
        <span className="px-2 py-1 rounded-full text-xs bg-emerald-900/30 text-emerald-300">
          {formatBytes(+disk.size)}
        </span>
      </td>
      <td className="py-3 px-4 text-sm text-gray-400">
        {formatDate(new Date(disk.created))}
      </td>
      <td className="py-3 px-4 text-sm text-gray-400">
        {formatDate(new Date(disk.modified))}
      </td>
    </tr>
  );
};

export default DiskListItem;
