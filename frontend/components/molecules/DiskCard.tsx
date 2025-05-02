import { HiOutlineClock, HiOutlineFolderOpen } from "react-icons/hi";

import { Disk } from "@/types/Disk";
import React from "react";
import { formatBytes } from "@/utils/formatBytes";

interface DiskCardProps {
  disk: Disk;
  isSelected: boolean;
  onClick: (path: string) => void;
}

const DiskCard: React.FC<DiskCardProps> = ({ disk, isSelected, onClick }) => {
  // Format date to a readable string
  const formatDate = (date: Date) => {
    return (
      date.toLocaleDateString() +
      " " +
      date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    );
  };

  return (
    <div
      onClick={() => onClick(disk.path)}
      className={`
        relative bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden
        transition-all duration-300 transform hover:scale-[1.02]
        ${
          isSelected
            ? "ring-2 ring-blue-500 shadow-lg shadow-blue-500/10"
            : "ring-1 ring-neutral-700 hover:ring-blue-400/30"
        }
        cursor-pointer p-5
      `}
    >
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-medium text-white">{disk.name}</h3>
        <span className="text-emerald-300 font-semibold text-sm px-2 py-1 rounded bg-emerald-900/30">
          {formatBytes(+disk.size)}
        </span>
      </div>

      <div className="space-y-2 text-sm">
        {/* Path */}
        <div className="flex items-center space-x-2 p-2 rounded-lg bg-neutral-700/30">
          <HiOutlineFolderOpen className="w-4 h-4 text-blue-400" />
          <span className="text-gray-300 text-xs font-mono truncate">
            {disk.path}
          </span>
        </div>

        {/* Dates */}
        <div className="flex items-center justify-between mt-4 text-xs text-gray-400">
          <div className="flex items-center">
            <HiOutlineClock className="w-3 h-3 mr-1" />
            <span>Created: {formatDate(new Date(disk.created))}</span>
          </div>
          <div className="flex items-center">
            <HiOutlineClock className="w-3 h-3 mr-1" />
            <span>Modified: {formatDate(new Date(disk.modified))}</span>
          </div>
        </div>
      </div>

      {/* Quick actions */}
      <div className="absolute opacity-0 group-hover:opacity-100 top-2 right-2 flex space-x-1">
        <button className="p-1 rounded-full bg-neutral-700 hover:bg-neutral-600">
          <span className="sr-only">Edit</span>
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
            <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
          </svg>
        </button>
      </div>
    </div>
  );
};

export default DiskCard;
