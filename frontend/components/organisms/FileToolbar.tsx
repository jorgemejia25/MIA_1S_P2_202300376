import {
  HiOutlineRefresh,
  HiOutlineViewGrid,
  HiOutlineViewList,
} from "react-icons/hi";

import React from "react";

interface FileToolbarProps {
  loading: boolean;
  viewMode: "grid" | "list";
  onRefresh: () => void;
  onViewModeChange: (mode: "grid" | "list") => void;
}

const FileToolbar: React.FC<FileToolbarProps> = ({
  loading,
  viewMode,
  onRefresh,
  onViewModeChange,
}) => {
  return (
    <div className="flex items-center space-x-2 ml-4">
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
  );
};

export default FileToolbar;
