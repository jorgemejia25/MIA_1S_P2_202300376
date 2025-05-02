import React from "react";
import RefreshButton from "../atoms/RefreshButton";
import ViewModeToggle from "../atoms/ViewModeToggle";

interface ToolBarProps {
  loading: boolean;
  viewMode: "grid" | "list";
  onRefresh: () => void;
  onViewModeChange: (mode: "grid" | "list") => void;
}

const ToolBar: React.FC<ToolBarProps> = ({
  loading,
  viewMode,
  onRefresh,
  onViewModeChange,
}) => {
  return (
    <div className="flex justify-between items-center mb-6">
      <RefreshButton onClick={onRefresh} loading={loading} />
      <ViewModeToggle viewMode={viewMode} onViewModeChange={onViewModeChange} />
    </div>
  );
};

export default ToolBar;
