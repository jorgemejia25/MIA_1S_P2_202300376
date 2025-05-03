import { FileContent } from "@/types/FileSystem";

/**
 * Lee el contenido de un archivo
 * @param diskPath Ruta del disco
 * @param partitionName Nombre de la partición
 * @param filePath Ruta del archivo
 * @returns Contenido del archivo o error
 */
export async function readFileContent(
  diskPath: string,
  partitionName: string,
  filePath: string
): Promise<FileContent> {
  try {
    console.log("Leyendo archivo:", filePath);
    console.log(diskPath, partitionName, filePath);

    const apiUrl = process.env.API_URL || "http://3.85.93.122:8080";
    const response = await fetch(`${apiUrl}/read-file`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        diskPath,
        partitionName,
        filePath,
      }),
      cache: "no-store",
    });

    const data = await response.json();

    if (!data.success) {
      return {
        success: false,
        message: data.message || "Error al leer el archivo",
        name: filePath,
        content: "",
      };
    }

    return {
      success: true,
      message: "Archivo leído correctamente",
      name: data.name || filePath,
      content: data.content || "",
    };
  } catch (error) {
    console.error("Error al leer el archivo:", error);
    return {
      success: false,
      message: "Error en la conexión con el servidor",
      name: filePath,
      content: "",
    };
  }
}
