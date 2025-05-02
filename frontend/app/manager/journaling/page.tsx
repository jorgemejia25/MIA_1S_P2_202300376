"use client";

import React, { useCallback, useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import { JournalEntry } from "@/components/molecules/JournalEntryCard";
import JournalingPageTemplate from "@/components/templates/JournalingPageTemplate";
import { getJournaling } from "@/actions/getJournaling";

const JournalingPage = () => {
  const searchParams = useSearchParams();
  const router = useRouter();
  const diskPath = searchParams.get("diskPath") || "";
  const partitionName = searchParams.get("partitionName") || "";

  // Estados
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [journalEntries, setJournalEntries] = useState<JournalEntry[]>([]);
  const [selectedEntry, setSelectedEntry] = useState<JournalEntry | null>(null);

  // Cargar los datos del journaling
  const loadJournalingData = useCallback(async () => {
    if (!diskPath || !partitionName) {
      setError("Se requiere especificar un disco y una partición");
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await getJournaling(diskPath, partitionName);

      if (response.success && response.entries) {
        setJournalEntries(response.entries);
      } else {
        setError(response.message || "Error al obtener el journaling");
      }
    } catch (err) {
      console.error("Error al cargar el journaling:", err);
      setError("Error al cargar los datos del journaling");
    } finally {
      setLoading(false);
    }
  }, [diskPath, partitionName]);

  // Cargar datos al iniciar
  useEffect(() => {
    loadJournalingData();
  }, [loadJournalingData]);

  // Manejadores de eventos
  const handleEntryClick = (entry: JournalEntry) => {
    setSelectedEntry(entry);
  };

  const handleCloseModal = () => {
    setSelectedEntry(null);
  };

  const handleRefresh = () => {
    loadJournalingData();
  };

  // Si no hay disco o partición especificados, redirigir a la página de archivos
  useEffect(() => {
    if (!diskPath || !partitionName) {
      router.replace("/manager/files");
    }
  }, [diskPath, partitionName, router]);

  return (
    <JournalingPageTemplate
      journalEntries={journalEntries}
      loading={loading}
      error={error}
      selectedEntry={selectedEntry}
      diskPath={diskPath}
      partitionName={partitionName}
      onEntryClick={handleEntryClick}
      onCloseModal={handleCloseModal}
      onRefresh={handleRefresh}
    />
  );
};

export default JournalingPage;
