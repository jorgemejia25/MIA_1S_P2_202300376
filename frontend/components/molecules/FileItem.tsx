"use client";

import {
  HiOutlineCalendar,
  HiOutlineDocumentText,
  HiOutlineFolderOpen,
} from "react-icons/hi";

import { FileInfo } from "@/types/FileSystem";
import { formatBytes } from "@/utils/formatBytes";

type FileItemProps = {
  file: FileInfo;
  onClick: (file: FileInfo) => void;
  viewMode: "grid" | "list";
  isSelected?: boolean;
};

const FileItem = ({
  file,
  onClick,
  viewMode,
  isSelected = false,
}: FileItemProps) => {
  const handleClick = () => {
    onClick(file);
  };

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString("es-ES", {
      day: "2-digit",
      month: "short",
      year: "numeric",
    });
  };

  // Determinar colores basados en el tipo de archivo
  let typeColor = "text-blue-400";
  let typeBackground = "bg-blue-900/30";

  if (file.type === "directory") {
    typeColor = "text-emerald-400";
    typeBackground = "bg-emerald-900/30";
  }

  if (viewMode === "grid") {
    return (
      <div
        className={`
          group flex flex-col items-center p-4 rounded-xl overflow-hidden
          bg-neutral-800/60 backdrop-blur-sm
          transition-all duration-300 transform hover:scale-[1.02] cursor-pointer
          ${
            isSelected
              ? "ring-2 ring-blue-500 shadow-lg shadow-blue-500/10"
              : "ring-1 ring-neutral-700 hover:ring-blue-400/30"
          }
        `}
        onClick={handleClick}
      >
        <div className="flex items-center justify-center w-14 h-14 mb-3 rounded-xl bg-neutral-800/50 group-hover:bg-neutral-700/50">
          {file.type === "directory" ? (
            <HiOutlineFolderOpen className="h-7 w-7 text-emerald-400" />
          ) : (
            <HiOutlineDocumentText className="h-7 w-7 text-gray-300" />
          )}
        </div>
        <div className="w-full text-center">
          <p
            className="font-medium text-white truncate max-w-full"
            title={file.name}
          >
            {file.name}
          </p>

          <div className="mt-2 space-y-1">
            {file.type === "directory" ? (
              <div
                className={`inline-block px-2 py-0.5 rounded-lg ${typeBackground}`}
              >
                <span className={`text-xs font-medium ${typeColor}`}>
                  Directorio
                </span>
              </div>
            ) : (
              <div className="flex flex-col items-center space-y-1">
                <div
                  className={`inline-block px-2 py-0.5 rounded-lg ${typeBackground}`}
                >
                  <span className={`text-xs font-medium ${typeColor}`}>
                    {formatBytes(file.size)}
                  </span>
                </div>
                <div className="flex items-center text-xs text-gray-400">
                  <HiOutlineCalendar className="w-3 h-3 mr-1" />
                  <span>{formatDate(file.modTime)}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  } else {
    return (
      <div
        className={`
          group flex items-center p-3 rounded-xl overflow-hidden
          bg-neutral-800/60 backdrop-blur-sm
          transition-all duration-300 cursor-pointer
          ${
            isSelected
              ? "ring-2 ring-blue-500 shadow-lg shadow-blue-500/10"
              : "ring-1 ring-neutral-700 hover:ring-blue-400/30"
          }
        `}
        onClick={handleClick}
      >
        <div
          className={`flex items-center justify-center w-10 h-10 mr-3 rounded-lg ${typeBackground}`}
        >
          {file.type === "directory" ? (
            <HiOutlineFolderOpen className="h-5 w-5 text-emerald-400" />
          ) : (
            <HiOutlineDocumentText className="h-5 w-5 text-gray-300" />
          )}
        </div>
        <div className="flex-grow min-w-0 mr-4">
          <p className="text-white truncate font-medium" title={file.name}>
            {file.name}
          </p>
          <p className="text-xs text-gray-400 md:hidden">
            {file.type === "directory" ? "Directorio" : formatBytes(file.size)}{" "}
            • {formatDate(file.modTime)}
          </p>
        </div>

        <div className="hidden md:flex items-center text-xs text-gray-400 w-28 justify-end">
          <HiOutlineCalendar className="w-3 h-3 mr-1" />
          <span>{formatDate(file.modTime)}</span>
        </div>

        {file.type !== "directory" && (
          <div className="hidden md:block text-xs font-medium w-20 text-right">
            <span
              className={`${typeColor} px-2 py-0.5 rounded-lg ${typeBackground}`}
            >
              {formatBytes(file.size)}
            </span>
          </div>
        )}

        {file.type === "directory" && (
          <div className="hidden md:block text-xs font-medium w-20 text-right">
            <span
              className={`${typeColor} px-2 py-0.5 rounded-lg ${typeBackground}`}
            >
              Directorio
            </span>
          </div>
        )}
      </div>
    );
  }
};

export default FileItem;
