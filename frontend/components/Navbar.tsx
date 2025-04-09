import React, { useRef } from "react";

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
    <nav className="absolute top-0 left-0 w-full flex justify-center items-center py-5">
      <div className="bg-neutral-800 text-white p-2 pl-5 gap-6 rounded-xl flex items-center justify-center cursor-pointer">
        <p className="font-semibold mr-5">ExtUsac</p>
        <p onClick={handleImportClick}>Importar</p>
        <input
          aria-label="Importar archivo"
          type="file"
          ref={fileInputRef}
          style={{ display: "none" }}
          onChange={handleFileChange}
          accept=".txt,.js,.ts,.json"
        />
        <p onClick={onClear}>Limpiar</p>
        <button
          type="button"
          className="bg-neutral-900 text-white p-2 px-3 rounded-lg flex items-center justify-center"
          onClick={onExecute}
        >
          Ejecutar
          <VscVmRunning className="inline ml-2" />
        </button>
      </div>
    </nav>
  );
};

export default Navbar;
