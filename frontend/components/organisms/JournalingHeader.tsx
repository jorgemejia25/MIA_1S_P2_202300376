import React from "react";
import RefreshButton from "@/components/atoms/RefreshButton";

interface JournalingHeaderProps {
  diskPath: string;
  partitionName: string;
  onRefresh: () => void;
}

const JournalingHeader: React.FC<JournalingHeaderProps> = ({
  diskPath,
  partitionName,
  onRefresh,
}) => {
  return (
    <div className="bg-neutral-800/50 backdrop-blur-sm rounded-xl p-6 border border-neutral-700/50 shadow-lg">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-5">
        <div>
          <h1 className="text-2xl font-medium text-white tracking-tight">
            Journaling
          </h1>
          <p className="text-gray-400 mt-1.5">
            Registro de operaciones del sistema de archivos
          </p>

          <div className="text-xs text-gray-500 mt-3 flex flex-wrap gap-3">
            <span className="inline-flex items-center">
              <svg
                className="h-3 w-3 mr-1.5"
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
              <span className="text-gray-300 font-mono">{diskPath}</span>
            </span>

            <span className="text-gray-500">•</span>

            <span className="inline-flex items-center">
              <svg
                className="h-3 w-3 mr-1.5"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                />
              </svg>
              <span className="text-gray-300 font-mono">{partitionName}</span>
            </span>
          </div>
        </div>

        <RefreshButton onRefresh={onRefresh} />
      </div>
    </div>
  );
};

export default JournalingHeader;
