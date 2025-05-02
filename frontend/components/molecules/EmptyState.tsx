import React from "react";

interface EmptyStateProps {
  title: string;
  description: string;
}

const EmptyState: React.FC<EmptyStateProps> = ({ title, description }) => {
  return (
    <div className="text-center py-16 bg-neutral-800/60 backdrop-blur-sm rounded-xl">
      <h3 className="text-xl font-medium text-gray-300 mb-2">{title}</h3>
      <p className="text-gray-400">{description}</p>
    </div>
  );
};

export default EmptyState;
