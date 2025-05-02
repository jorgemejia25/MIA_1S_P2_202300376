"use client";

import { DirectoryContent, FileContent, FileInfo } from "@/types/FileSystem";
import React, { useEffect, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";

import FilesPageTemplate from "@/components/templates/FilesPageTemplate";
import { listDirectory } from "@/actions/listDirectory";
import { readFileContent } from "@/actions/readFileContent";

const FilesPage = () => {
  const router = useRouter();
  const searchParams = useSearchParams();
  const initialPath = searchParams.get("path") || "/";
  const diskPath = searchParams.get("diskPath") || "";
  const partitionName = searchParams.get("partitionName") || "";

  const [currentPath, setCurrentPath] = useState(initialPath);
  const [directoryContent, setDirectoryContent] =
    useState<DirectoryContent | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

  // Estado para manejar el archivo seleccionado y su contenido
  const [selectedFile, setSelectedFile] = useState<FileInfo | null>(null);
  const [fileContent, setFileContent] = useState<FileContent | null>(null);
  const [loadingContent, setLoadingContent] = useState(false);

  // Dividir la ruta en segmentos para la barra de navegación
  const pathSegments = currentPath
    .split("/")
    .filter((segment) => segment !== "");

  // Construir las rutas para cada segmento
  const breadcrumbItems = pathSegments.map((segment, index) => {
    const path = "/" + pathSegments.slice(0, index + 1).join("/");
    return { name: segment, path };
  });

  // Añadir el elemento raíz al inicio
  breadcrumbItems.unshift({ name: "Raíz", path: "/" });

  const handlePathChange = (newPath: string) => {
    setCurrentPath(newPath);
    loadDirectory(diskPath, partitionName, newPath);

    // Cerrar el visor de archivos si está abierto
    setSelectedFile(null);
    setFileContent(null);

    // Actualizar la URL con el nuevo path
    const params = new URLSearchParams(searchParams);
    params.set("path", newPath);
    router.replace(`/manager/files?${params.toString()}`);
  };

  const handleFileClick = async (file: FileInfo) => {
    if (file.type === "directory") {
      // Construir la nueva ruta para navegación
      const newPath = currentPath.endsWith("/")
        ? `${currentPath}${file.name}`
        : `${currentPath}/${file.name}`;
      handlePathChange(newPath);
    } else {
      // Es un archivo, mostrar su contenido
      setSelectedFile(file);
      loadFileContent(file);
    }
  };

  const loadFileContent = async (file: FileInfo) => {
    if (!diskPath || !partitionName) {
      setError("Se requiere especificar un disco y una partición");
      return;
    }

    setLoadingContent(true);

    try {
      // Construir la ruta del archivo
      const filePath =
        currentPath === "/" ? `/${file.name}` : `${currentPath}/${file.name}`;

      const content = await readFileContent(diskPath, partitionName, filePath);
      setFileContent(content);

      if (!content.success) {
        setError(content.message || "Error al cargar el contenido del archivo");
      }
    } catch (err) {
      console.error("Error al cargar el contenido del archivo:", err);
      setError("Error al cargar el contenido del archivo");
    } finally {
      setLoadingContent(false);
    }
  };

  const closeFileViewer = () => {
    setSelectedFile(null);
    setFileContent(null);
  };

  const handleRefresh = async () => {
    loadDirectory(diskPath, partitionName, currentPath);
  };

  const handleViewModeChange = (mode: "grid" | "list") => {
    setViewMode(mode);
  };

  // Función para manejar la navegación al Journaling
  const handleJournalingClick = () => {
    if (!diskPath || !partitionName) {
      setError(
        "Se requiere especificar un disco y una partición para ver el journaling"
      );
      return;
    }

    // Navegar a la página de journaling con los parámetros necesarios
    router.push(
      `/manager/journaling?diskPath=${diskPath}&partitionName=${partitionName}`
    );
  };

  const loadDirectory = async (
    disk: string,
    partition: string,
    path: string
  ) => {
    if (!disk || !partition) {
      setError("Se requiere especificar un disco y una partición");
      setLoading(false);
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const response = await listDirectory(disk, partition, path);

      if (response.success && response.content) {
        setDirectoryContent(response.content);
      } else {
        setError(response.message || "Error al cargar el directorio");
        setDirectoryContent(null);
      }
    } catch (err) {
      setError("Error al cargar los archivos");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  // Cargar directorio al iniciar
  useEffect(() => {
    loadDirectory(diskPath, partitionName, currentPath);
  }, [diskPath, partitionName, currentPath]);

  return (
    <FilesPageTemplate
      currentPath={currentPath}
      directoryContent={directoryContent}
      loading={loading}
      error={error}
      viewMode={viewMode}
      selectedFile={selectedFile}
      fileContent={fileContent}
      loadingContent={loadingContent}
      pathSegments={breadcrumbItems}
      onPathChange={handlePathChange}
      onFileClick={handleFileClick}
      onCloseFileViewer={closeFileViewer}
      onRefresh={handleRefresh}
      onViewModeChange={handleViewModeChange}
      onJournalingClick={handleJournalingClick}
    />
  );
};

export default FilesPage;
