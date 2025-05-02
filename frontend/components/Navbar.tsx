import React, { useRef } from "react";

import Link from "next/link";
import { MdClear } from "react-icons/md";
import { RiPhoneFindLine } from "react-icons/ri";
import { VscVmRunning } from "react-icons/vsc";

interface NavbarProps {
  onExecute: () => void;
  onClear: () => void;
  onImport: (fileContent: string) => void;
}

const Navbar: React.FC<NavbarProps> = ({ onExecute, onClear, onImport }) => {
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleImportClick = () => {
    fileInputRef.current?.click();
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (event) => {
      const content = event.target?.result as string;
      onImport(content);
    };
    reader.readAsText(file);
  };

  return (
    <nav className="absolute top-0 left-0 w-full flex justify-center items-center py-3">
      <div className="bg-neutral-900 border border-neutral-700 text-white px-6 py-2 rounded-lg flex items-center shadow-lg">
        <Link href="/" className="font-bold text-teal-400 mr-6">
          ExtUsac
        </Link>

        <div className="flex items-center gap-5">
          <button
            onClick={handleImportClick}
            className="flex items-center gap-1 text-neutral-300 hover:text-teal-400 transition-colors"
          >
            <RiPhoneFindLine />
            <span className="text-sm">Importar</span>
          </button>

          <input
            aria-label="Importar archivo"
            type="file"
            ref={fileInputRef}
            className="hidden"
            onChange={handleFileChange}
            accept=".txt,.js,.ts,.json"
          />

          <button
            onClick={onClear}
            className="flex items-center gap-1 text-neutral-300 hover:text-teal-400 transition-colors"
          >
            <MdClear />
            <span className="text-sm">Limpiar</span>
          </button>

          <div className="h-4 w-px bg-neutral-700 mx-1"></div>

          <Link
            href="/manager/disks"
            className="text-sm text-neutral-300 hover:text-teal-400 transition-colors"
          >
            Sistema de archivos
          </Link>

          <Link
            href="/auth"
            className="text-sm text-neutral-300 hover:text-teal-400 transition-colors"
          >
            Login
          </Link>
        </div>

        <button
          type="button"
          className="ml-6 bg-teal-600 hover:bg-teal-700 text-white text-sm py-1.5 px-3 rounded-md flex items-center transition-colors"
          onClick={onExecute}
        >
          Ejecutar
          <VscVmRunning className="ml-1.5" />
        </button>
      </div>
    </nav>
  );
};

export default Navbar;
