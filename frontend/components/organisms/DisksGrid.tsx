import { Disk } from "@/types/Disk";
import DiskCard from "../molecules/DiskCard";
import React from "react";

interface DisksGridProps {
  disks: Disk[];
  selectedDisk: string | null;
  onDiskSelect: (path: string) => void;
}

const DisksGrid: React.FC<DisksGridProps> = ({
  disks,
  selectedDisk,
  onDiskSelect,
}) => {
  return (
    <div className="grid md:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4 gap-8">
      {disks.map((disk) => (
        <DiskCard
          key={disk.path}
          disk={disk}
          isSelected={selectedDisk === disk.name}
          onClick={onDiskSelect}
        />
      ))}
    </div>
  );
};

export default DisksGrid;
