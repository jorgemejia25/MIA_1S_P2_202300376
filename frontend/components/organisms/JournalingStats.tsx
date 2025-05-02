import { HiOutlineClock } from "react-icons/hi";
import { JournalEntry } from "@/components/molecules/JournalEntryCard";
import JournalingStatCard from "../molecules/JournalingStatCard";
import OperationIcon from "../atoms/OperationIcon";
import React from "react";
import { getOperationConfig } from "@/utils/operationStyles";

interface JournalingStatsProps {
  journalEntries: JournalEntry[];
}

const JournalingStats: React.FC<JournalingStatsProps> = ({
  journalEntries,
}) => {
  // No renderizar si no hay entradas
  if (!journalEntries.length) return null;

  // Obtener la última entrada del journal
  const lastEntry = journalEntries[journalEntries.length - 1];
  const lastEntryConfig = getOperationConfig(lastEntry?.operation || "default");

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
    <div className="grid grid-cols-1 md:grid-cols-3 gap-5">
      {/* Tarjeta de total de operaciones */}
      <JournalingStatCard
        title="Total de operaciones"
        value={journalEntries.length}
        icon={
          <svg
            className="w-5 h-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"
            />
          </svg>
        }
        iconBg="bg-blue-600/10"
        iconClass="text-blue-400"
      />

      {/* Tarjeta de última operación */}
      <JournalingStatCard
        title="Última operación"
        value={lastEntry?.operation || "N/A"}
        icon={<OperationIcon operation={lastEntry?.operation || ""} />}
        iconBg={lastEntryConfig.iconBg}
        iconClass={lastEntryConfig.iconClass}
      />

      {/* Tarjeta de último registro */}
      <JournalingStatCard
        title="Último registro"
        value={lastEntry ? formatDate(lastEntry.date) : "N/A"}
        icon={<HiOutlineClock className="w-5 h-5" />}
        iconBg="bg-purple-600/10"
        iconClass="text-purple-400"
      />
    </div>
  );
};

export default JournalingStats;
