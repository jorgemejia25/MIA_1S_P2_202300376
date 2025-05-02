"use client";

import React, { useActionState, useEffect, useState } from "react";

import { login } from "@/actions/login";
import { redirect } from "next/navigation";

const Login = () => {
  const [partition, setPartition] = useState("");
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const [state, formAction] = useActionState(login, {
    success: undefined,
    msg: undefined,
  });

  useEffect(() => {
    if (state.success) {
      redirect("/?login=true");
    }
  }, [state]);

  return (
    <section className="w-full h-screen flex">
      <div className="h-full w-1/2 flex items-center justify-center relative">
        <div className="absolute top-10 left-10 flex items-center">
          <div className="w-10 h-10 rounded-xl bg-white text-black flex items-center justify-center font-bold ">
            JM
          </div>
          <span className="ml-2 text-white font-semibold">
            Manejo de Archivos
          </span>
        </div>
        <div>
          <div className="text-center text-xl font-bold mb-2">
            Inicia sesión con tu cuenta
          </div>
          <div className="text-center text-gray-500">
            Ingresa tus datos para continuar
          </div>
          <form className="flex flex-col gap-4 mt-10 w-96" action={formAction}>
            <div>
              <label className="text-xs mb-2 block font-semibold">
                ID Partición
              </label>
              <input
                type="text"
                name="partition"
                placeholder="ID partición"
                className="w-full p-3 rounded-lg bg-transparent border-neutral-800 border text-white focus:outline-none focus:ring-2 focus:ring-teal-500"
                value={partition}
                onChange={(e) => setPartition(e.target.value)}
              />
            </div>
            <div>
              <label className="text-xs mb-2 block font-semibold">
                Usuario
              </label>
              <input
                type="text"
                name="username"
                placeholder="Usuario"
                className="w-full p-3 rounded-lg bg-transparent border-neutral-800 border text-white focus:outline-none focus:ring-2 focus:ring-teal-500"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
              />
            </div>
            <div>
              <label className="text-xs mb-2 block font-semibold">
                Contraseña
              </label>
              <input
                name="password"
                type="password"
                placeholder="Contraseña"
                className="w-full p-3 rounded-lg bg-transparent border-neutral-800 border text-white focus:outline-none focus:ring-2 focus:ring-teal-500"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>

            {state.msg && (
              <div className="text-red-500 text-xs mt-2">{state.msg}</div>
            )}

            <div>
              <button
                type="submit"
                className="w-full p-3 mt-5 rounded-lg bg-white text-black font-semibold hover:bg-teal-50 transition duration-200"
              >
                Iniciar sesión
              </button>
            </div>
          </form>
          <footer className="absolute bottom-5 left-1/2 transform -translate-x-1/2 text-center text-gray-500">
            Desarrollado por Jorge Mejía
          </footer>
        </div>
      </div>
      <div className="h-full w-1/2 bg-neutral-900"></div>
    </section>
  );
};

export default Login;
