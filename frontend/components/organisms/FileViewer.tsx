import { FileContent, FileInfo } from "@/types/FileSystem";
import {
  HiOutlineCode,
  HiOutlineDocumentText,
  HiOutlineExternalLink,
  HiOutlineInformationCircle,
  HiOutlineX,
} from "react-icons/hi";

import LoadingSpinner from "../atoms/LoadingSpinner";
import React from "react";
import { formatBytes } from "@/utils/formatBytes";

interface FileViewerProps {
  selectedFile: FileInfo;
  fileContent: FileContent | null;
  loadingContent: boolean;
  onClose: () => void;
}

const FileViewer: React.FC<FileViewerProps> = ({
  selectedFile,
  fileContent,
  loadingContent,
  onClose,
}) => {
  return (
    <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden ring-1 ring-neutral-700">
      {/* Cabecera del archivo con información */}
      <div className="flex items-center justify-between border-b border-neutral-700 p-4">
        <div className="flex items-center">
          <div className="flex items-center justify-center w-10 h-10 mr-3 rounded-lg bg-blue-900/30">
            <HiOutlineDocumentText className="h-5 w-5 text-blue-400" />
          </div>
          <div>
            <h3 className="text-white font-medium">{selectedFile?.name}</h3>
            <div className="flex items-center mt-1 text-xs text-gray-400">
              <span className="flex items-center mr-3">
                <HiOutlineCode className="mr-1 h-3.5 w-3.5" />
                {formatBytes(selectedFile?.size || 0)}
              </span>
              <span className="hidden md:flex items-center">
                <HiOutlineInformationCircle className="mr-1 h-3.5 w-3.5" />
                {new Date(selectedFile?.modTime || "").toLocaleDateString(
                  "es-ES",
                  {
                    day: "2-digit",
                    month: "short",
                    year: "numeric",
                  }
                )}
              </span>
            </div>
          </div>
        </div>
        <div className="flex items-center">
          <button
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-neutral-700/50 text-gray-400 hover:text-white transition-colors"
          >
            <HiOutlineX className="h-5 w-5" />
            <span className="sr-only">Cerrar</span>
          </button>
        </div>
      </div>

      {/* Contenido del archivo */}
      <div className="p-4 flex-grow">
        {loadingContent ? (
          <div className="flex items-center justify-center h-96">
            <LoadingSpinner />
          </div>
        ) : (
          fileContent &&
          (fileContent.success ? (
            <div className="relative h-full overflow-auto">
              <div className="h-full p-6 font-mono text-sm bg-neutral-900/80 rounded-lg shadow-inner border border-neutral-700/50">
                <pre className="text-emerald-100/90 leading-relaxed tracking-wide whitespace-pre-wrap">
                  {typeof fileContent.content === "string"
                    ? fileContent.content
                    : JSON.stringify(fileContent.content, null, 4)}
                </pre>
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center p-8 h-96 text-center">
              <div className="bg-red-900/20 p-6 rounded-xl border border-red-800/30 max-w-md">
                <HiOutlineExternalLink className="h-12 w-12 text-red-400/80 mx-auto mb-4" />
                <h4 className="text-red-300 text-lg font-medium mb-2">
                  Error al cargar el archivo
                </h4>
                <p className="text-neutral-300/80">
                  No se pudo cargar el contenido del archivo.
                </p>
                <p className="mt-2 text-sm text-neutral-400 bg-neutral-800/50 p-3 rounded border border-neutral-700/50">
                  {fileContent.message}
                </p>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default FileViewer;
