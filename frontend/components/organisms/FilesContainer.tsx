import { DirectoryContent, FileInfo } from "@/types/FileSystem";

import FilesGrid from "./FilesGrid";
import FilesList from "./FilesList";
import React from "react";

interface FilesContainerProps {
  directoryContent: DirectoryContent | null;
  viewMode: "grid" | "list";
  loading: boolean;
  onFileClick: (file: FileInfo) => void;
  onRefresh: () => void;
}

const FilesContainer: React.FC<FilesContainerProps> = ({
  directoryContent,
  viewMode,
  loading,
  onFileClick,
  onRefresh,
}) => {
  if (!loading && directoryContent) {
    if (directoryContent.files.length > 0) {
      return viewMode === "grid" ? (
        <FilesGrid files={directoryContent.files} onFileClick={onFileClick} />
      ) : (
        <FilesList files={directoryContent.files} onFileClick={onFileClick} />
      );
    } else {
      // Estado vacío
      return (
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
            onClick={onRefresh}
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
      );
    }
  }

  // Estado no disponible
  return (
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
        Selecciona un disco y una partición para explorar archivos
      </p>
    </div>
  );
};

export default FilesContainer;
