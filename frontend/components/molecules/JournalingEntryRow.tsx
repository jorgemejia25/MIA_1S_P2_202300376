import { HiOutlineClock } from "react-icons/hi";
import { JournalEntry } from "@/components/molecules/JournalEntryCard";
import OperationIcon from "../atoms/OperationIcon";
import React from "react";
import { getOperationConfig } from "@/utils/operationStyles";

interface JournalingEntryRowProps {
  entry: JournalEntry;
  onClick: () => void;
}

const JournalingEntryRow: React.FC<JournalingEntryRowProps> = ({
  entry,
  onClick,
}) => {
  const opConfig = getOperationConfig(entry.operation);

  const formatDate = (date: Date): string => {
    return (
      date.toLocaleDateString() +
      " " +
      date.toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      })
    );
  };

  return (
    <div
      className={`grid grid-cols-12 px-6 py-4 transition-colors cursor-pointer ${opConfig.hoverBg}`}
      onClick={onClick}
    >
      {/* Columna de fecha */}
      <div className="col-span-3 flex items-center text-sm text-gray-300">
        <HiOutlineClock className="w-4 h-4 mr-2.5 text-gray-500 flex-shrink-0" />
        <span className="truncate">{formatDate(entry.date)}</span>
      </div>

      {/* Columna de operación */}
      <div className="col-span-2">
        <div
          className={`inline-flex items-center space-x-2 px-2.5 py-1.5 rounded-md text-xs font-medium ${opConfig.bg} ${opConfig.text} ${opConfig.border}`}
        >
          <span className={opConfig.iconClass}>
            <OperationIcon operation={entry.operation} />
          </span>
          <span>{entry.operation}</span>
        </div>
      </div>

      {/* Columna de ruta */}
      <div className="col-span-4 flex items-center">
        <div className="text-sm text-gray-300 truncate font-mono">
          {entry.path}
        </div>
      </div>

      {/* Columna de contenido */}
      <div className="col-span-3 text-sm text-gray-400 truncate">
        {entry.content ? (
          entry.content
        ) : (
          <span className="text-gray-500 italic">Sin contenido</span>
        )}
      </div>
    </div>
  );
};

export default JournalingEntryRow;
