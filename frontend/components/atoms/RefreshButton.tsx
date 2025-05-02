import { HiOutlineRefresh } from "react-icons/hi";
import React from "react";

interface RefreshButtonProps {
  onClick: () => void;
  loading: boolean;
}

const RefreshButton: React.FC<RefreshButtonProps> = ({ onClick, loading }) => {
  return (
    <button
      onClick={onClick}
      className="flex items-center p-2 px-3 rounded-md bg-blue-600 hover:bg-blue-700 transition-colors"
      disabled={loading}
    >
      <HiOutlineRefresh
        className={`w-4 h-4 mr-2 ${loading ? "animate-spin" : ""}`}
      />
      <span>{loading ? "Cargando..." : "Actualizar"}</span>
    </button>
  );
};

export default RefreshButton;
