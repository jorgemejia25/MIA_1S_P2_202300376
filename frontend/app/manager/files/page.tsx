"use client";

import { HiOutlineFolderOpen, HiOutlineInformationCircle } from "react-icons/hi";
import React, { useState } from "react";

import ErrorMessage from "@/components/atoms/ErrorMessage";
import LoadingSpinner from "@/components/atoms/LoadingSpinner";
import PageHeader from "@/components/organisms/PageHeader";
import ToolBar from "@/components/organisms/ToolBar";
import { useSearchParams } from "next/navigation";

const FilesPage = () => {
  const searchParams = useSearchParams();
  const initialPath = searchParams.get("path") || "/";

  const [currentPath, setCurrentPath] = useState(initialPath);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

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
    // Aquí posteriormente se cargarían los archivos y carpetas de la nueva ruta
  };

  const handleRefresh = async () => {
    setLoading(true);
    try {
      // Aquí se llamaría a una función para cargar los archivos
      // Por ahora solo simulamos un retraso
      await new Promise((resolve) => setTimeout(resolve, 500));
    } catch (err) {
      setError("Error al cargar los archivos");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleViewModeChange = (mode: "grid" | "list") => {
    setViewMode(mode);
  };

  return (
    <div className="max-w-7xl mx-auto">
      <PageHeader
        title="Explorador de Archivos"
        description="Navega por el sistema de archivos y gestiona tus directorios y archivos."
      />

      <ToolBar
        loading={loading}
        viewMode={viewMode}
        onRefresh={handleRefresh}
        onViewModeChange={handleViewModeChange}
      />

      <ErrorMessage message={error} />

      {/* Barra de navegación de directorio con la temática de DiskCard */}
      <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden p-4 mb-4 ring-1 ring-neutral-700 hover:ring-emerald-400/30 transition-all">
        <div className="flex flex-wrap items-center">
          {breadcrumbItems.map((item, index) => (
            <React.Fragment key={item.path}>
              <button
                onClick={() => handlePathChange(item.path)}
                className={`text-sm hover:bg-neutral-700/30 px-3 py-1.5 rounded transition-colors ${
                  index === breadcrumbItems.length - 1
                    ? "bg-emerald-900/30 text-emerald-300 font-medium"
                    : "text-gray-300"
                }`}
              >
                {index === 0 ? (
                  <span className="flex items-center">
                    <HiOutlineFolderOpen className="h-4 w-4 mr-1.5 text-emerald-400" />
                    Raíz
                  </span>
                ) : (
                  item.name
                )}
              </button>
              {index < breadcrumbItems.length - 1 && (
                <span className="mx-2 text-gray-600">
                  <svg 
                    xmlns="http://www.w3.org/2000/svg" 
                    className="h-4 w-4" 
                    fill="none" 
                    viewBox="0 0 24 24" 
                    stroke="currentColor"
                  >
                    <path 
                      strokeLinecap="round" 
                      strokeLinejoin="round" 
                      strokeWidth={2} 
                      d="M9 5l7 7-7 7" 
                    />
                  </svg>
                </span>
              )}
            </React.Fragment>
          ))}
        </div>

        <div className="mt-3 border-t border-neutral-700 pt-2">
          <div className="flex items-center p-2 rounded-lg bg-neutral-700/30 text-xs text-gray-400">
            <HiOutlineInformationCircle className="h-4 w-4 mr-1.5 text-gray-400" />
            <span className="font-medium mr-1.5">Ruta actual:</span>
            <code className="font-mono text-xs text-gray-300 truncate">
              {currentPath}
            </code>
          </div>
        </div>
      </div>

      {loading && <LoadingSpinner />}

      {/* Contenedor de archivos con la temática de DiskCard */}
      <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden p-5 ring-1 ring-neutral-700">
        {!loading && (
          <div className="text-center p-10">
            <svg 
              xmlns="http://www.w3.org/2000/svg" 
              className="mx-auto h-14 w-14 text-neutral-700" 
              fill="none" 
              viewBox="0 0 24 24" 
              stroke="currentColor"
            >
              <path 
                strokeLinecap="round" 
                strokeLinejoin="round" 
                strokeWidth={1.5} 
                d="M5 19a2 2 0 01-2-2V7a2 2 0 012-2h4l2 2h4a2 2 0 012 2v1M5 19h14a2 2 0 002-2v-5a2 2 0 00-2-2H9a2 2 0 00-2 2v5a2 2 0 01-2 2z" 
              />
            </svg>
            <p className="mt-4 text-gray-400">
              No hay archivos para mostrar en este directorio
            </p>
            <button 
              onClick={handleRefresh}
              className="mt-5 inline-flex items-center px-4 py-2 rounded-md text-white bg-neutral-700 hover:bg-neutral-600 transition-colors"
            >
              <svg 
                xmlns="http://www.w3.org/2000/svg" 
                className="h-4 w-4 mr-2" 
                fill="none" 
                viewBox="0 0 24 24" 
                stroke="currentColor"
              >
                <path 
                  strokeLinecap="round" 
                  strokeLinejoin="round" 
                  strokeWidth={2} 
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" 
                />
              </svg>
              Refrescar directorio
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default FilesPage;
