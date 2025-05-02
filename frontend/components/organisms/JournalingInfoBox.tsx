import { HiOutlineInformationCircle } from "react-icons/hi";
import React from "react";

const JournalingInfoBox: React.FC = () => {
  return (
    <div className="bg-blue-900/5 border border-blue-800/20 rounded-lg p-5 shadow-lg">
      <div className="flex items-start">
        <div className="flex-shrink-0 p-2 rounded-full bg-blue-500/10 mr-4">
          <HiOutlineInformationCircle className="h-5 w-5 text-blue-400" />
        </div>

        <div>
          <h4 className="text-blue-300 font-medium mb-1.5">
            ¿Qué es el journaling?
          </h4>

          <p className="text-sm text-blue-200/70 leading-relaxed">
            El journaling registra todas las operaciones realizadas en el
            sistema de archivos para permitir la recuperación en caso de fallos
            inesperados. Las operaciones se almacenan en el inicio de la
            partición, justo después del SuperBloque, creando así un historial
            detallado que puede utilizarse para reconstruir el estado del
            sistema de archivos en caso de una interrupción abrupta.
          </p>
        </div>
      </div>
    </div>
  );
};

export default JournalingInfoBox;
