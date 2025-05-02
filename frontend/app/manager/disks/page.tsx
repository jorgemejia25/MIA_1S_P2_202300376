"use client";

import React, { useEffect, useState } from "react";

import { Disk } from "@/types/Disk";
import DisksPageTemplate from "@/components/templates/DisksPageTemplate";
import { listDisks } from "@/actions/listDisks";
import { useRouter } from "next/navigation";

const DisksPage = () => {
  const [disks, setDisks] = useState<Disk[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedDisk, setSelectedDisk] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const router = useRouter();

  // Cargar los discos desde el backend
  useEffect(() => {
    const fetchDisks = async () => {
      try {
        setLoading(true);
        const response = await listDisks();

        console.log(response);

        if (response.success && response.disks) {
          setDisks(response.disks);
          setError(null);
        } else {
          setError(response.msg || "Error al cargar los discos");
          setDisks([]);
        }
      } catch (err) {
        setError("Error inesperado al cargar los discos");
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchDisks();
  }, []);

  // Función para recargar la lista de discos
  const handleRefresh = async () => {
    try {
      setLoading(true);
      const response = await listDisks();

      if (response.success && response.disks) {
        setDisks(response.disks);
        setError(null);
      } else {
        setError(response.msg || "Error al recargar los discos");
      }
    } catch (err) {
      setError("Error inesperado al recargar los discos");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Redirigir a la página de particiones cuando se selecciona un disco
  const handleDiskSelection = (diskPath: string) => {
    setSelectedDisk(diskPath);
    // Redirección a la página de particiones con el path como parámetro de consulta
    router.push(`/manager/partitions?path=${encodeURIComponent(diskPath)}`);
  };

  // Función para cambiar el modo de visualización
  const handleViewModeChange = (mode: "grid" | "list") => {
    setViewMode(mode);
  };

  return (
    <DisksPageTemplate
      disks={disks}
      loading={loading}
      error={error}
      selectedDisk={selectedDisk}
      viewMode={viewMode}
      onRefresh={handleRefresh}
      onViewModeChange={handleViewModeChange}
      onDiskSelect={handleDiskSelection}
    />
  );
};

export default DisksPage;
