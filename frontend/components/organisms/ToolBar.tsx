"use client";

import RefreshButton from "@/components/atoms/RefreshButton";
import ViewModeToggle from "@/components/atoms/ViewModeToggle";

type ToolBarProps = {
  loading?: boolean;
  viewMode: "grid" | "list";
  onRefresh: () => void;
  onViewModeChange: (mode: "grid" | "list") => void;
};

const ToolBar = ({
  loading = false,
  viewMode,
  onRefresh,
  onViewModeChange,
}: ToolBarProps) => {
  return (
    <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl p-3 mb-4 flex justify-between items-center ring-1 ring-neutral-700">
      <div className="flex items-center space-x-2">
        <RefreshButton loading={loading} onRefresh={onRefresh} />
      </div>
      <div className="flex items-center">
        <ViewModeToggle
          viewMode={viewMode}
          onViewModeChange={onViewModeChange}
        />
      </div>
    </div>
  );
};

export default ToolBar;
