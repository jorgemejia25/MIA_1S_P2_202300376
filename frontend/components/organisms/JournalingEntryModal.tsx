import { JournalEntry } from "@/components/molecules/JournalEntryCard";
import JournalEntryCard from "@/components/molecules/JournalEntryCard";
import React from "react";

interface JournalingEntryModalProps {
  selectedEntry: JournalEntry | null;
  onClose: () => void;
}

const JournalingEntryModal: React.FC<JournalingEntryModalProps> = ({
  selectedEntry,
  onClose,
}) => {
  if (!selectedEntry) return null;

  return (
    <div
      className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50 p-6"
      onClick={onClose}
    >
      <div
        className="max-w-2xl w-full transform transition-all"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="relative bg-neutral-800/90 rounded-lg overflow-hidden border border-neutral-700/60 shadow-2xl">
          {/* Botón para cerrar el modal */}
          <button
            onClick={onClose}
            className="absolute top-4 right-4 text-gray-400 hover:text-white p-2 rounded-full hover:bg-neutral-700/40 transition-colors"
            aria-label="Cerrar"
          >
            <svg
              className="w-5 h-5"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M6 18L18 6M6 6l12 12"
              />
            </svg>
          </button>

          <JournalEntryCard entry={selectedEntry} />
        </div>
      </div>
    </div>
  );
};

export default JournalingEntryModal;
