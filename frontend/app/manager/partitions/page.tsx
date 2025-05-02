"use client";

import React, { useCallback, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { Partition } from "@/types/Partition";
import PartitionsPageTemplate from "@/components/templates/PartitionsPageTemplate";
import { listPartitions } from "@/actions/listPartitions";

const PartitionsPage = () => {
  const searchParams = useSearchParams();
  const router = useRouter();
  const diskPath = searchParams.get("path");

  const [partitions, setPartitions] = useState<Partition[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedPartition, setSelectedPartition] = useState<string | null>(
    null
  );
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

  const fetchPartitions = useCallback(async () => {
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
  }, [diskPath]);

  useEffect(() => {
    // Si tenemos el path del disco, cargar sus particiones
    if (diskPath) {
      fetchPartitions();
    } else {
      setError("No se ha seleccionado ningún disco");
      setLoading(false);
    }
  }, [diskPath, fetchPartitions]);

  // Función para recargar la lista de particiones
  const handleRefresh = async () => {
    if (diskPath) {
      await fetchPartitions();
    }
  };

  // Función para manejar la selección de una partición
  const handlePartitionSelection = (name: string) => {
    setSelectedPartition(name);
    // Redireccionar a la página de archivos con los parámetros de disco y partición
    router.push(
      `/manager/files?diskPath=${encodeURIComponent(
        diskPath || ""
      )}&partitionName=${encodeURIComponent(name)}`
    );
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
