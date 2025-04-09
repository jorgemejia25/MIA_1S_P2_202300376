"use client";

import CodeEditor from "@/components/CodeEditor";
import Navbar from "@/components/Navbar";
import { sendCommand } from "@/actions/sendCommand";
import { useState } from "react";

export default function Home() {
  const [code, setCode] = useState("");
  const [output, setOutput] = useState("// Salida del código aquí...");

  const handleCodeChange = (newCode: string) => {
    setCode(newCode);
  };

  const handleExecute = async () => {
    const result = await sendCommand(code);
    setOutput(result);
  };

  const handleClear = () => {
    setCode("");
    setOutput("// Salida del código aquí...");
  };

  const handleImport = (fileContent: string) => {
    setCode(fileContent);
  };

  return (
    <div className="flex flex-col items-start justify-between min-h-screen px-20 pt-28 pb-6 gap-16 font-[family-name:var(--font-geist-sans)] bg-neutral-900">
      <Navbar
        onExecute={handleExecute}
        onClear={handleClear}
        onImport={handleImport}
      />
      <main className="flex flex-col gap-5 w-full">
        <h1 className="text-6xl font-bold text-gray-200">Proyecto 1</h1>
        <h1 className="text-2xl text-gray-400">
          Manejo e implementación de archivos
        </h1>
        <span className="mt-5"></span>
        <CodeEditor
          code={code}
          output={output}
          onCodeChange={handleCodeChange}
        />
      </main>
      <footer className="text-center text-gray-500 w-full">
        Jorge Mejía - 202300376
      </footer>
    </div>
  );
}
