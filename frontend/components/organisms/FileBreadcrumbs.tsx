import {
  HiOutlineFolderOpen,
  HiOutlineInformationCircle,
  HiOutlineRefresh,
  HiOutlineViewGrid,
  HiOutlineViewList,
} from "react-icons/hi";

import React from "react";

interface PathSegment {
  name: string;
  path: string;
}

interface FileBreadcrumbsProps {
  pathSegments: PathSegment[];
  currentPath: string;
  onPathChange: (newPath: string) => void;
  loading: boolean;
  viewMode: "grid" | "list";
  onRefresh: () => void;
  onViewModeChange: (mode: "grid" | "list") => void;
  onJournalingClick?: () => void;
}

const FileBreadcrumbs: React.FC<FileBreadcrumbsProps> = ({
  pathSegments,
  currentPath,
  onPathChange,
  loading,
  viewMode,
  onRefresh,
  onViewModeChange,
  onJournalingClick,
}) => {
  return (
    <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden mb-6 ring-1 ring-neutral-700 hover:ring-emerald-400/30 transition-all">
      {/* Parte superior con breadcrumbs y botones */}
      <div className="flex items-center justify-between p-3 border-b border-neutral-700/50">
        <div className="flex flex-wrap items-center flex-grow overflow-x-auto scrollbar-hide">
          {pathSegments.map((item, index) => (
            <React.Fragment key={item.path}>
              <button
                onClick={() => onPathChange(item.path)}
                className={`text-sm hover:bg-neutral-700/30 px-3 py-1.5 rounded transition-colors ${
                  index === pathSegments.length - 1
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

              {index < pathSegments.length - 1 && (
                <span className="mx-1 text-gray-600">
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
        <div className="flex items-center space-x-2">
          {/* Botón de Journaling */}
          <button
            onClick={onJournalingClick}
            className="p-2 rounded-lg transition-colors px-3 bg-neutral-700/50 hover:bg-neutral-700/30 text-gray-300 hover:text-emerald-300"
          >
            <span>Journaling</span>
          </button>

          {/* Botón de refrescar */}
          <button
            onClick={onRefresh}
            disabled={loading}
            className={`p-2 rounded-lg transition-colors ${
              loading
                ? "opacity-50 cursor-not-allowed bg-neutral-700/50"
                : "hover:bg-neutral-700/50 text-emerald-400 hover:text-emerald-300"
            }`}
          >
            <HiOutlineRefresh
              className={`w-5 h-5 ${loading ? "animate-spin" : ""}`}
            />
            <span className="sr-only">Refrescar</span>
          </button>

          {/* Botones de vista */}
          <div className="flex bg-neutral-700/30 rounded-lg p-0.5">
            <button
              onClick={() => onViewModeChange("grid")}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === "grid"
                  ? "bg-neutral-600 text-white"
                  : "text-gray-400 hover:text-white"
              }`}
            >
              <HiOutlineViewGrid className="w-4 h-4" />
              <span className="sr-only">Vista de cuadrícula</span>
            </button>
            <button
              onClick={() => onViewModeChange("list")}
              className={`p-1.5 rounded-md transition-colors ${
                viewMode === "list"
                  ? "bg-neutral-600 text-white"
                  : "text-gray-400 hover:text-white"
              }`}
            >
              <HiOutlineViewList className="w-4 h-4" />
              <span className="sr-only">Vista de lista</span>
            </button>
          </div>
        </div>
      </div>

      {/* Barra inferior con ruta actual */}
      <div className="flex items-center px-4 py-2 bg-neutral-800/80">
        <div className="flex items-center text-xs text-gray-400 w-full">
          <HiOutlineInformationCircle className="h-4 w-4 mr-1.5 text-gray-400" />
          <span className="font-medium mr-1.5">Ruta:</span>
          <code className="font-mono text-xs text-gray-300 truncate overflow-x-auto scrollbar-hide max-w-full">
            {currentPath}
          </code>
        </div>
      </div>
    </div>
  );
};

export default FileBreadcrumbs;
