"use client";

import { HiOutlineRefresh } from "react-icons/hi";

type RefreshButtonProps = {
  loading?: boolean;
  onRefresh: () => void;
};

const RefreshButton = ({ loading = false, onRefresh }: RefreshButtonProps) => {
  return (
    <button
      onClick={onRefresh}
      disabled={loading}
      className={`flex items-center justify-center p-2 rounded-md transition-colors 
        ${loading 
          ? "bg-neutral-700/50 text-gray-500 cursor-not-allowed" 
          : "bg-neutral-800/80 text-gray-300 hover:text-emerald-300 hover:bg-neutral-700/80"
        }`}
      title="Refrescar"
    >
      <HiOutlineRefresh 
        className={`h-5 w-5 ${loading ? "animate-spin" : ""}`} 
      />
    </button>
  );
};

export default RefreshButton;
