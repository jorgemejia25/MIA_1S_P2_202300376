import { JournalEntry } from "@/components/molecules/JournalEntryCard";

interface JournalingResponse {
  success: boolean;
  message?: string;
  entries?: JournalEntry[];
}

export async function getJournaling(
  diskPath: string,
  partitionName: string
): Promise<JournalingResponse> {
  try {
    const apiUrl =
      process.env.NEXT_PUBLIC_API_URL || "http://54.196.151.70:8080";

    // Construir la URL con los parámetros de consulta
    const url = `${apiUrl}/journaling?diskPath=${encodeURIComponent(
      diskPath
    )}&partitionName=${encodeURIComponent(partitionName)}`;

    // Realizar la petición al backend
    const response = await fetch(url);

    if (!response.ok) {
      throw new Error(
        `Error de red: ${response.status} ${response.statusText}`
      );
    }

    // Parsear la respuesta como JSON
    const data: JournalingResponse = await response.json();

    // Si hay entradas, convertir las fechas de string a objetos Date
    if (data.success && data.entries) {
      data.entries = data.entries.map((entry) => ({
        ...entry,
        date: new Date(entry.date),
      }));
    }

    return data;
  } catch (error) {
    console.error("Error al obtener el journaling:", error);
    return {
      success: false,
      message:
        "Error al obtener el journaling: " +
        (error instanceof Error ? error.message : "Error desconocido"),
    };
  }
}
