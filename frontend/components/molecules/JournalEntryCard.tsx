import {
  HiOutlineClock,
  HiOutlineDocumentText,
  HiOutlineFolder,
} from "react-icons/hi";

import React from "react";

// Tipo para las entradas de journaling
export interface JournalEntry {
  operation: string;
  path: string;
  content: string;
  date: Date;
}

interface JournalEntryCardProps {
  entry: JournalEntry;
}

const JournalEntryCard: React.FC<JournalEntryCardProps> = ({ entry }) => {
  // Función para determinar el icono basado en la operación
  const getOperationIcon = (operation: string) => {
    switch (operation) {
      case "mkdir":
        return <HiOutlineFolder className="w-6 h-6 text-blue-400" />;
      case "mkfile":
      case "append":
      case "rename":
        return <HiOutlineDocumentText className="w-6 h-6 text-emerald-400" />;
      case "chmod":
        return (
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="w-6 h-6 text-amber-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
            />
          </svg>
        );
      case "rm":
        return (
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="w-6 h-6 text-red-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
            />
          </svg>
        );
      default:
        return <HiOutlineDocumentText className="w-6 h-6 text-gray-400" />;
    }
  };

  // Función para formatear la fecha
  const formatDate = (date: Date) => {
    return (
      date.toLocaleDateString() +
      " " +
      date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    );
  };

  // Función para obtener el color de la operación
  const getOperationColor = (operation: string) => {
    switch (operation) {
      case "mkdir":
        return "bg-blue-900/20 text-blue-300 border-blue-700/30";
      case "mkfile":
        return "bg-emerald-900/20 text-emerald-300 border-emerald-700/30";
      case "chmod":
        return "bg-amber-900/20 text-amber-300 border-amber-700/30";
      case "append":
        return "bg-purple-900/20 text-purple-300 border-purple-700/30";
      case "rm":
        return "bg-red-900/20 text-red-300 border-red-700/30";
      case "rename":
        return "bg-cyan-900/20 text-cyan-300 border-cyan-700/30";
      default:
        return "bg-gray-700/20 text-gray-300 border-gray-600";
    }
  };

  // Describir la operación en lenguaje natural
  const getOperationDescription = (operation: string) => {
    switch (operation) {
      case "mkdir":
        return "Creación de directorio";
      case "mkfile":
        return "Creación de archivo";
      case "chmod":
        return "Cambio de permisos";
      case "append":
        return "Añadir contenido";
      case "rm":
        return "Eliminar archivo";
      case "rename":
        return "Renombrar archivo";
      default:
        return operation;
    }
  };

  // Función para determinar el gradiente de fondo basado en la operación
  const getHeaderGradient = (operation: string) => {
    switch (operation) {
      case "mkdir":
        return "from-blue-600/20 to-blue-900/20";
      case "mkfile":
        return "from-emerald-600/20 to-emerald-900/20";
      case "chmod":
        return "from-amber-600/20 to-amber-900/20";
      case "append":
        return "from-purple-600/20 to-purple-900/20";
      case "rm":
        return "from-red-600/20 to-red-900/20";
      case "rename":
        return "from-cyan-600/20 to-cyan-900/20";
      default:
        return "from-gray-700/20 to-gray-900/20";
    }
  };

  return (
    <div className="overflow-hidden transform transition-all">
      {/* Encabezado */}
      <div
        className={`bg-gradient-to-r ${getHeaderGradient(
          entry.operation
        )} p-6 rounded-t-xl backdrop-blur-sm`}
      >
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <div
              className={`p-3 rounded-full ${
                entry.operation === "mkdir"
                  ? "bg-blue-600/30"
                  : entry.operation === "mkfile"
                  ? "bg-emerald-600/30"
                  : entry.operation === "chmod"
                  ? "bg-amber-600/30"
                  : entry.operation === "append"
                  ? "bg-purple-600/30"
                  : entry.operation === "rm"
                  ? "bg-red-600/30"
                  : entry.operation === "rename"
                  ? "bg-cyan-600/30"
                  : "bg-gray-600/30"
              }`}
            >
              {getOperationIcon(entry.operation)}
            </div>
            <div>
              <h3 className="text-xl font-medium bg-clip-text text-transparent bg-gradient-to-r from-white to-gray-300">
                {getOperationDescription(entry.operation)}
              </h3>
              <div className="flex items-center mt-1 text-xs text-gray-400">
                <HiOutlineClock className="w-3.5 h-3.5 mr-1.5" />
                <span>{formatDate(entry.date)}</span>
              </div>
            </div>
          </div>
          <div
            className={`inline-flex items-center px-3 py-1.5 rounded-full text-xs font-medium ${getOperationColor(
              entry.operation
            )}`}
          >
            {entry.operation}
          </div>
        </div>
      </div>

      {/* Contenido */}
      <div className="p-6 space-y-5 bg-neutral-800/60 backdrop-blur-sm rounded-b-xl">
        {/* Ruta */}
        <div>
          <div className="flex items-center mb-2">
            <svg
              className="h-4 w-4 text-gray-400 mr-2"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
              />
            </svg>
            <span className="text-sm font-medium text-gray-300">
              Ruta del recurso
            </span>
          </div>
          <div className="bg-neutral-700/30 rounded-lg p-3.5 text-sm text-gray-300 font-mono border border-neutral-600/30 shadow-inner">
            {entry.path}
          </div>
        </div>

        {/* Contenido si existe */}
        {entry.content && (
          <div>
            <div className="flex items-center mb-2">
              <svg
                className="h-4 w-4 text-gray-400 mr-2"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
                />
              </svg>
              <span className="text-sm font-medium text-gray-300">
                Contenido
              </span>
            </div>
            <div className="bg-neutral-700/30 rounded-lg p-3.5 text-sm text-gray-300 font-mono border border-neutral-600/30 shadow-inner">
              {entry.content}
            </div>
          </div>
        )}

        {/* Detalles técnicos */}
        <div className="border-t border-neutral-700/50 pt-5 mt-5">
          <div className="flex items-center mb-3">
            <svg
              className="h-4 w-4 text-gray-400 mr-2"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
              />
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
              />
            </svg>
            <span className="text-sm font-medium text-gray-300">
              Detalles técnicos
            </span>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div className="bg-neutral-700/20 backdrop-blur-sm rounded-lg p-4 border border-neutral-700/30">
              <div className="text-xs text-gray-500 mb-1 uppercase tracking-wide">
                Fecha de operación
              </div>
              <div className="text-sm text-gray-300 font-mono">
                {entry.date.toISOString()}
              </div>
            </div>
            <div className="bg-neutral-700/20 backdrop-blur-sm rounded-lg p-4 border border-neutral-700/30">
              <div className="text-xs text-gray-500 mb-1 uppercase tracking-wide">
                Tipo de operación
              </div>
              <div className="text-sm text-gray-300 flex items-center">
                <span
                  className={`inline-block w-2 h-2 rounded-full mr-2 ${
                    entry.operation === "mkdir"
                      ? "bg-blue-400"
                      : entry.operation === "mkfile"
                      ? "bg-emerald-400"
                      : entry.operation === "chmod"
                      ? "bg-amber-400"
                      : entry.operation === "append"
                      ? "bg-purple-400"
                      : entry.operation === "rm"
                      ? "bg-red-400"
                      : entry.operation === "rename"
                      ? "bg-cyan-400"
                      : "bg-gray-400"
                  }`}
                ></span>
                {getOperationDescription(entry.operation)}
              </div>
            </div>
            <div className="bg-neutral-700/20 backdrop-blur-sm rounded-lg p-4 border border-neutral-700/30">
              <div className="text-xs text-gray-500 mb-1 uppercase tracking-wide">
                Timestamp
              </div>
              <div className="text-sm text-gray-300 font-mono">
                {entry.date.getTime()}
              </div>
            </div>
            <div className="bg-neutral-700/20 backdrop-blur-sm rounded-lg p-4 border border-neutral-700/30">
              <div className="text-xs text-gray-500 mb-1 uppercase tracking-wide">
                Código de operación
              </div>
              <div className="text-sm text-gray-300 font-mono">
                {entry.operation}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default JournalEntryCard;
