"use client";

import { useEffect, useState } from "react";

import CodeEditor from "@/components/CodeEditor";
import Navbar from "@/components/Navbar";
import { logout } from "@/actions/logout";
import { sendCommand } from "@/actions/sendCommand";
import { useSearchParams } from "next/navigation";

export default function Home() {
  const [code, setCode] = useState("");
  const [output, setOutput] = useState("// Salida del código aquí...");
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const searchParams = useSearchParams();

  useEffect(() => {
    const loginParam = searchParams.get("login");
    if (loginParam) {
      setIsLoggedIn(true);
    }
  }, [searchParams]);

  const handleCodeChange = (newCode: string) => {
    setCode(newCode);
  };

  const handleExecute = async () => {
    setIsLoading(true);
    const result = await sendCommand(code);
    setOutput(result);
    setIsLoading(false);
  };

  const handleClear = () => {
    setCode("");
    setOutput("// Salida del código aquí...");
  };

  const handleImport = (fileContent: string) => {
    setCode(fileContent);
  };

  const handleLogout = async () => {
    const result = await logout();
    if (result.success) {
      setIsLoggedIn(false);
    }
  };

  return (
    <div className="flex flex-col min-h-screen  text-white">
      {/* Navigation Bar */}
      <div className="fixed top-0 left-0 right-0 z-10 bg-black/90 backdrop-blur-sm border-b border-gray-800">
        <Navbar
          onExecute={handleExecute}
          onClear={handleClear}
          onImport={handleImport}
        />
      </div>

      {/* Main Content Area */}
      <main className="flex-1 pt-28 px-6 md:px-12 lg:px-20 pb-6 max-w-7xl mx-auto w-full">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl md:text-5xl lg:text-6xl font-bold text-gray-100 tracking-tight">
            Proyecto 2
          </h1>
          <p className="text-xl text-gray-400 mt-2">
            Manejo e implementación de archivos
          </p>
        </div>

        {/* Login Success Message */}
        {isLoggedIn && (
          <div className="bg-emerald-950/60 border border-emerald-500/50 rounded-lg p-4 mb-6 flex items-center justify-between shadow-lg transition-all duration-300 ease-in-out">
            <div className="flex items-center">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-5 w-5 mr-3 text-emerald-400"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5 13l4 4L19 7"
                />
              </svg>
              <p className="font-medium text-emerald-300">
                Sesión iniciada correctamente
              </p>
            </div>
            <button
              onClick={handleLogout}
              className="px-3 py-1 rounded-md text-emerald-300 hover:text-white hover:bg-emerald-800/50 transition-colors duration-200"
            >
              Cerrar sesión
            </button>
          </div>
        )}

        {/* Code Editor */}
        <div className="rounded-xl overflow-hidden shadow-2xl">
          <CodeEditor
            code={code}
            output={output}
            onCodeChange={handleCodeChange}
            isLoading={isLoading}
          />
        </div>
      </main>

      {/* Footer */}
      <footer className="py-4 text-center text-sm text-gray-500 border-t border-gray-800/50">
        Jorge Mejía - 202300376
      </footer>
    </div>
  );
}
