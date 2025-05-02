import React from "react";
import { HiOutlineDocumentText } from "react-icons/hi";
import { JournalEntry } from "@/components/molecules/JournalEntryCard";
import LoadingSpinner from "@/components/atoms/LoadingSpinner";
import ErrorMessage from "@/components/atoms/ErrorMessage";
import EmptyState from "@/components/molecules/EmptyState";
import JournalingHeader from "@/components/organisms/JournalingHeader";
import JournalingStats from "@/components/organisms/JournalingStats";
import JournalingEntryTable from "@/components/organisms/JournalingEntryTable";
import JournalingInfoBox from "@/components/organisms/JournalingInfoBox";
import JournalingEntryModal from "@/components/organisms/JournalingEntryModal";

interface JournalingPageTemplateProps {
  journalEntries: JournalEntry[];
  loading: boolean;
  error: string | null;
  selectedEntry: JournalEntry | null;
  diskPath: string;
  partitionName: string;
  onEntryClick: (entry: JournalEntry) => void;
  onCloseModal: () => void;
  onRefresh: () => void;
}

const JournalingPageTemplate: React.FC<JournalingPageTemplateProps> = ({
  journalEntries,
  loading,
  error,
  selectedEntry,
  diskPath,
  partitionName,
  onEntryClick,
  onCloseModal,
  onRefresh,
}) => {
  return (
    <div className="p-6 min-h-screen ">
      <div className="max-w-6xl mx-auto space-y-6">
        {/* Header */}
        <JournalingHeader
          diskPath={diskPath}
          partitionName={partitionName}
          onRefresh={onRefresh}
        />

        {/* Error Message */}
        {error && <ErrorMessage message={error} />}

        {/* Contenido principal */}
        {loading ? (
          <div className="py-12">
            <LoadingSpinner
              center
              message="Cargando entradas de journaling..."
            />
          </div>
        ) : journalEntries.length === 0 ? (
          <div className="py-16">
            <EmptyState
              icon={<HiOutlineDocumentText className="w-16 h-16" />}
              title="No hay entradas de journaling"
              description="Esta partición no tiene operaciones registradas en el journaling o no tiene habilitado ext3."
            />
          </div>
        ) : (
          <div className="space-y-6">
            {/* Tarjetas de estadísticas */}
            <JournalingStats journalEntries={journalEntries} />

            {/* Lista de entradas */}
            <JournalingEntryTable
              journalEntries={journalEntries}
              onEntryClick={onEntryClick}
            />

            {/* Nota informativa */}
            <JournalingInfoBox />
          </div>
        )}

        {/* Modal para detalles */}
        <JournalingEntryModal
          selectedEntry={selectedEntry}
          onClose={onCloseModal}
        />
      </div>
    </div>
  );
};

export default JournalingPageTemplate;
