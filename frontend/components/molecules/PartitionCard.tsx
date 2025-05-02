import { HiOutlineCheckCircle, HiOutlineDatabase } from "react-icons/hi";

import { Partition } from "@/types/Partition";
import React from "react";
import { formatBytes } from "@/utils/formatBytes";

interface PartitionCardProps {
  partition: Partition;
  isSelected: boolean;
  onClick: (name: string) => void;
}

const PartitionCard: React.FC<PartitionCardProps> = ({
  partition,
  isSelected,
  onClick,
}) => {
  let typeColor = "text-blue-400";
  let typeBackground = "bg-blue-900/30";

  switch (partition.type.toLowerCase()) {
    case "primaria":
    case "primary":
    case "p":
      typeColor = "text-green-400";
      typeBackground = "bg-green-900/30";
      break;
    case "extendida":
    case "extended":
    case "e":
      typeColor = "text-yellow-400";
      typeBackground = "bg-yellow-900/30";
      break;
    case "lógica":
    case "logical":
    case "l":
      typeColor = "text-purple-400";
      typeBackground = "bg-purple-900/30";
      break;
  }

  return (
    <div
      onClick={() => onClick(partition.name)}
      className={`
        relative bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden
        transition-all duration-300 transform hover:scale-[1.02]
        ${
          isSelected
            ? "ring-2 ring-blue-500 shadow-lg shadow-blue-500/10"
            : "ring-1 ring-neutral-700 hover:ring-blue-400/30"
        }
        group cursor-pointer p-5
      `}
    >
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-medium text-white">{partition.name}</h3>
        <span className="text-emerald-300 font-semibold text-sm px-2 py-1 rounded bg-emerald-900/30">
          {formatBytes(partition.size)}
        </span>
      </div>

      <div className="space-y-2 text-sm">
        {/* Type */}
        <div className={`inline-block px-3 py-1 rounded-lg ${typeBackground}`}>
          <span className={`font-medium ${typeColor}`}>{partition.type}</span>
        </div>

        {/* Status */}
        <div className="flex items-center space-x-2 mt-3">
          <div
            className={`w-3 h-3 rounded-full ${
              partition.status === "active" ? "bg-green-500" : "bg-gray-500"
            }`}
          />
          <span className="text-gray-400">
            {partition.status === "active" ? "Activo" : "Inactivo"}
          </span>
        </div>

        {/* Mounted */}
        <div className="flex items-center mt-3">
          <HiOutlineDatabase className="w-4 h-4 text-blue-400 mr-2" />
          {partition.isMounted ? (
            <div className="flex items-center">
              <HiOutlineCheckCircle className="w-4 h-4 text-green-400 mr-1" />
              <span className="text-green-300">ID: {partition.mountId}</span>
            </div>
          ) : (
            <span className="text-gray-400">No montada</span>
          )}
        </div>
      </div>

      {/* Quick actions (opcional) */}
      <div className="absolute opacity-0 group-hover:opacity-100 top-2 right-2 flex space-x-1">
        <button className="p-1 rounded-full bg-neutral-700 hover:bg-neutral-600 transition-colors">
          <span className="sr-only">Ver detalles</span>
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
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="16" x2="12" y2="12"></line>
            <line x1="12" y1="8" x2="12.01" y2="8"></line>
          </svg>
        </button>
      </div>
    </div>
  );
};

export default PartitionCard;
