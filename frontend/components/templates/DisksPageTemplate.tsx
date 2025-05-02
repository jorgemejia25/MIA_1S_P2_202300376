import React from "react";
import { Disk } from "@/types/Disk";
import PageHeader from "../organisms/PageHeader";
import ToolBar from "../organisms/ToolBar";
import ErrorMessage from "../atoms/ErrorMessage";
import LoadingSpinner from "../atoms/LoadingSpinner";
import EmptyState from "../molecules/EmptyState";
import DisksGrid from "../organisms/DisksGrid";
import DisksList from "../organisms/DisksList";

interface DisksPageTemplateProps {
  disks: Disk[];
  loading: boolean;
  error: string | null;
  selectedDisk: string | null;
  viewMode: "grid" | "list";
  onRefresh: () => void;
  onViewModeChange: (mode: "grid" | "list") => void;
  onDiskSelect: (path: string) => void;
}

const DisksPageTemplate: React.FC<DisksPageTemplateProps> = ({
  disks,
  loading,
  error,
  selectedDisk,
  viewMode,
  onRefresh,
  onViewModeChange,
  onDiskSelect,
}) => {
  return (
    <>
      <div className="max-w-7xl mx-auto">
        <PageHeader
          title="Selector de Discos"
          description="Selecciona un disco para ver sus particiones y detalles. Puedes cambiar entre vista de cuadrícula y lista."
        />

        <ToolBar
          loading={loading}
          viewMode={viewMode}
          onRefresh={onRefresh}
          onViewModeChange={onViewModeChange}
        />

        <ErrorMessage message={error} />

        {loading && disks.length === 0 && <LoadingSpinner />}

        {!loading && disks.length === 0 && !error && (
          <EmptyState
            title="No hay discos disponibles"
            description="Crea un nuevo disco para empezar"
          />
        )}

        {!loading && disks.length > 0 && viewMode === "grid" && (
          <DisksGrid
            disks={disks}
            selectedDisk={selectedDisk}
            onDiskSelect={onDiskSelect}
          />
        )}

        {!loading && disks.length > 0 && viewMode === "list" && (
          <DisksList
            disks={disks}
            selectedDisk={selectedDisk}
            onDiskSelect={onDiskSelect}
          />
        )}
      </div>
    </>
  );
};

export default DisksPageTemplate;
