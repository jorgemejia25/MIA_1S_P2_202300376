import React, { ReactNode } from "react";

interface JournalingStatCardProps {
  title: string;
  value: string | number;
  icon: ReactNode;
  iconBg?: string;
  iconClass?: string;
}

const JournalingStatCard: React.FC<JournalingStatCardProps> = ({
  title,
  value,
  icon,
  iconBg = "bg-blue-600/10",
  iconClass = "text-blue-400",
}) => (
  <div className="bg-neutral-800/40 rounded-lg p-5 border border-neutral-700/40 transition-all duration-200 hover:bg-neutral-800/60 hover:border-neutral-600/40 hover:shadow-md">
    <div className="flex items-center">
      <div className={`p-3 rounded-md ${iconBg} mr-4`}>
        <span className={iconClass}>{icon}</span>
      </div>

      <div>
        <div className="text-gray-400 text-xs font-medium">{title}</div>
        <div className="text-xl font-medium text-white mt-1">{value}</div>
      </div>
    </div>
  </div>
);

export default JournalingStatCard;
