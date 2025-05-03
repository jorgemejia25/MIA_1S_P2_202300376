import { DirectoryLsResponse } from "@/types/FileSystem";

export async function listDirectory(
  disk: string,
  partition: string,
  path: string = "/"
): Promise<DirectoryLsResponse> {
  try {
    console.log("EJECUTANDO DISK", disk);

    // Construir la URL con parámetros de consulta
    const apiUrl = process.env.API_URL || "http://3.85.93.122:8080";
    const url = new URL(`${apiUrl}/directory`);
    url.searchParams.append("disk", disk);
    url.searchParams.append("partition", partition);
    url.searchParams.append("path", path);

    // Realizar la petición
    const response = await fetch(url.toString(), {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error(`Error al obtener el directorio: ${response.statusText}`);
    }

    const data = (await response.json()) as DirectoryLsResponse;
    return data;
  } catch (error) {
    console.error("Error en listDirectory:", error);
    return {
      success: false,
      message:
        error instanceof Error
          ? error.message
          : "Error desconocido al listar el directorio",
    };
  }
}
