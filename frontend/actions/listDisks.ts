import { Disk } from "../types/Disk";

interface ListDisksResponse {
  success: boolean;
  msg?: string;
  disks?: Disk[];
}

/**
 * Obtiene la lista de discos desde el backend
 * @returns Promise con la respuesta que contiene la lista de discos
 */
export const listDisks = async (): Promise<ListDisksResponse> => {
  try {
    const apiUrl =
      process.env.NEXT_PUBLIC_API_URL || "http://54.196.151.70:8080";
    const response = await fetch(`${apiUrl}/disks`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      // Utilizamos next: { revalidate: 0 } para evitar la caché y obtener siempre la lista actualizada
      next: { revalidate: 0 },
    });

    console.log(response);

    if (!response.ok) {
      throw new Error(`Error HTTP: ${response.status}`);
    }

    const data: ListDisksResponse = await response.json();

    // Transformar las fechas de string a Date
    if (data.disks) {
      data.disks = data.disks.map((disk) => ({
        ...disk,
        Created: new Date(disk.created),
        Modified: new Date(disk.modified),
      }));
    }

    return data;
  } catch (error) {
    console.error("Error al obtener la lista de discos:", error);
    return {
      success: false,
      msg: "Error de conexión con el servidor",
      disks: [],
    };
  }
};
