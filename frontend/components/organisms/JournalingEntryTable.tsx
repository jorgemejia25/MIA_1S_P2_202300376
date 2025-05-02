import { JournalEntry } from "@/components/molecules/JournalEntryCard";
import JournalingEntryRow from "../molecules/JournalingEntryRow";
import React from "react";

interface JournalingEntryTableProps {
  journalEntries: JournalEntry[];
  onEntryClick: (entry: JournalEntry) => void;
}

const JournalingEntryTable: React.FC<JournalingEntryTableProps> = ({
  journalEntries,
  onEntryClick,
}) => {
  return (
    <div className="bg-neutral-800/30 rounded-lg overflow-hidden border border-neutral-700/40 shadow-lg">
      {/* Encabezado de la tabla */}
      <div className="bg-neutral-800/60 py-4 px-6 border-b border-neutral-700/60">
        <h3 className="text-gray-200 font-medium">Historial de operaciones</h3>
      </div>

      {/* Títulos de columnas */}
      <div className="grid grid-cols-12 px-6 py-3 bg-neutral-800/60 border-b border-neutral-700/40 text-xs font-medium text-gray-400 uppercase tracking-wider">
        <div className="col-span-3">Fecha</div>
        <div className="col-span-2">Operación</div>
        <div className="col-span-4">Ruta</div>
        <div className="col-span-3">Contenido</div>
      </div>

      {/* Filas de datos */}
      <div className="divide-y divide-neutral-700/20 max-h-[calc(100vh-470px)] overflow-y-auto">
        {journalEntries.map((entry, index) => (
          <JournalingEntryRow
            key={index}
            entry={entry}
            onClick={() => onEntryClick(entry)}
          />
        ))}
      </div>
    </div>
  );
};

export default JournalingEntryTable;
