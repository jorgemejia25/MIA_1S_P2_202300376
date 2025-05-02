"use client";

import { HiOutlineViewGrid, HiOutlineViewList } from "react-icons/hi";

type ViewModeToggleProps = {
  viewMode: "grid" | "list";
  onViewModeChange: (mode: "grid" | "list") => void;
};

const ViewModeToggle = ({
  viewMode,
  onViewModeChange,
}: ViewModeToggleProps) => {
  return (
    <div className="flex items-center space-x-1 bg-neutral-800/80 rounded-lg p-1">
      <button
        onClick={() => onViewModeChange("grid")}
        className={`p-1.5 rounded-md transition-colors ${
          viewMode === "grid"
            ? "bg-emerald-900/40 text-emerald-300"
            : "text-gray-400 hover:text-gray-200 hover:bg-neutral-700/50"
        }`}
        title="Vista en cuadrícula"
      >
        <HiOutlineViewGrid className="h-5 w-5" />
      </button>
      <button
        onClick={() => onViewModeChange("list")}
        className={`p-1.5 rounded-md transition-colors ${
          viewMode === "list"
            ? "bg-emerald-900/40 text-emerald-300"
            : "text-gray-400 hover:text-gray-200 hover:bg-neutral-700/50"
        }`}
        title="Vista en lista"
      >
        <HiOutlineViewList className="h-5 w-5" />
      </button>
    </div>
  );
};

export default ViewModeToggle;
