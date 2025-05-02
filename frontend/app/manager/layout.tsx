import { Suspense } from "react";

export default function ManagerLayout(
  props: Readonly<{
    children: React.ReactNode;
  }>
) {
  return (
    <div className="min-h-screen text-gray-100 p-6">
      <div className="max-w-7xl mx-auto">
        <Suspense
          fallback={<div className="text-center p-10">Cargando...</div>}
        >
          {props.children}
        </Suspense>
      </div>
    </div>
  );
}
