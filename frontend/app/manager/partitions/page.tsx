"use client";

import React, { useEffect, useState } from "react";

import { Partition } from "@/types/Partition";
import PartitionsPageTemplate from "@/components/templates/PartitionsPageTemplate";
import { listPartitions } from "@/actions/listPartitions";
import { useSearchParams } from "next/navigation";

const PartitionsPage = () => {
  const searchParams = useSearchParams();
  const diskPath = searchParams.get("path");

  const [partitions, setPartitions] = useState<Partition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedPartition, setSelectedPartition] = useState<string | null>(
    null
  );
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

  useEffect(() => {
    // Si tenemos el path del disco, cargar sus particiones
    if (diskPath) {
      fetchPartitions();
    } else {
      setError("No se ha seleccionado ningún disco");
      setLoading(false);
    }
  }, [diskPath]);

  const fetchPartitions = async () => {
    try {
      setLoading(true);
      const response = await listPartitions(diskPath!);

      if (response.success && response.partitions) {
        setPartitions(response.partitions.partitions);
        setError(null);
      } else {
        setError(response.msg || "Error al cargar las particiones");
        setPartitions([]);
      }
    } catch (err) {
      setError("Error inesperado al cargar las particiones");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Función para recargar la lista de particiones
  const handleRefresh = async () => {
    if (diskPath) {
      await fetchPartitions();
    }
  };

  // Función para manejar la selección de una partición
  const handlePartitionSelection = (name: string) => {
    setSelectedPartition(name);
    // Aquí podríamos añadir funcionalidad adicional al seleccionar una partición
    // como mostrar detalles, opciones para formatear, etc.
  };

  // Función para cambiar el modo de visualización
  const handleViewModeChange = (mode: "grid" | "list") => {
    setViewMode(mode);
  };

  return (
    <PartitionsPageTemplate
      partitions={partitions}
      loading={loading}
      error={error}
      selectedPartition={selectedPartition}
      viewMode={viewMode}
      diskPath={diskPath}
      onRefresh={handleRefresh}
      onViewModeChange={handleViewModeChange}
      onPartitionSelect={handlePartitionSelection}
    />
  );
};

export default PartitionsPage;
