import React from "react";
import { Partition } from "@/types/Partition";
import PageHeader from "../organisms/PageHeader";
import ToolBar from "../organisms/ToolBar";
import ErrorMessage from "../atoms/ErrorMessage";
import LoadingSpinner from "../atoms/LoadingSpinner";
import EmptyState from "../molecules/EmptyState";
import PartitionsGrid from "../organisms/PartitionsGrid";
import PartitionsList from "../organisms/PartitionsList";

interface PartitionsPageTemplateProps {
  partitions: Partition[];
  loading: boolean;
  error: string | null;
  selectedPartition: string | null;
  viewMode: "grid" | "list";
  diskPath: string | null;
  onRefresh: () => void;
  onViewModeChange: (mode: "grid" | "list") => void;
  onPartitionSelect: (name: string) => void;
}

const PartitionsPageTemplate: React.FC<PartitionsPageTemplateProps> = ({
  partitions,
  loading,
  error,
  selectedPartition,
  viewMode,
  diskPath,
  onRefresh,
  onViewModeChange,
  onPartitionSelect,
}) => {
  return (
    < >
      <div className="max-w-7xl mx-auto">
        <PageHeader
          title="Particiones"
          description={
            diskPath
              ? `Particiones del disco: ${diskPath}`
              : "Selecciona un disco para ver sus particiones"
          }
        />

        <ToolBar
          loading={loading}
          viewMode={viewMode}
          onRefresh={onRefresh}
          onViewModeChange={onViewModeChange}
        />

        <ErrorMessage message={error} />

        {loading && partitions && partitions.length === 0 && <LoadingSpinner />}

        {!loading && partitions && partitions.length === 0 && !error && (
          <EmptyState
            title="No hay particiones disponibles"
            description={
              diskPath
                ? "Este disco no tiene particiones o no se pudieron cargar"
                : "Selecciona un disco para ver sus particiones"
            }
          />
        )}

        {!loading &&
          partitions &&
          partitions.length > 0 &&
          viewMode === "grid" && (
            <PartitionsGrid
              partitions={partitions}
              selectedPartition={selectedPartition}
              onPartitionSelect={onPartitionSelect}
            />
          )}

        {!loading &&
          partitions &&
          partitions.length > 0 &&
          viewMode === "list" && (
            <PartitionsList
              partitions={partitions}
              selectedPartition={selectedPartition}
              onPartitionSelect={onPartitionSelect}
            />
          )}
      </div>
    </>
  );
};

export default PartitionsPageTemplate;
