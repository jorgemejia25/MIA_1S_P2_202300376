import { Disk } from "@/types/Disk";
import DiskListItem from "../molecules/DiskListItem";
import React from "react";

interface DisksListProps {
  disks: Disk[];
  selectedDisk: string | null;
  onDiskSelect: (path: string) => void;
}

const DisksList: React.FC<DisksListProps> = ({
  disks,
  selectedDisk,
  onDiskSelect,
}) => {
  return (
    <div className="bg-neutral-800/60 backdrop-blur-sm rounded-xl overflow-hidden ring-1 ring-neutral-700">
      <table className="w-full">
        <thead>
          <tr className="border-b border-neutral-700">
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Nombre
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Ruta
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Tamaño
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Creado
            </th>
            <th className="text-left py-3 px-4 text-sm font-medium text-gray-400">
              Modificado
            </th>
          </tr>
        </thead>
        <tbody>
          {disks.map((disk) => (
            <DiskListItem
              key={disk.path}
              disk={disk}
              isSelected={selectedDisk === disk.name}
              onClick={onDiskSelect}
            />
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default DisksList;
