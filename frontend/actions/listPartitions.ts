"use server";

import { Partition } from "@/types/Partition";

type ListPartitionsResponse = {
  success: boolean;
  msg?: string;
  partitions?: {
    logicalPartitions: Partition[];
    partitions: Partition[];
  };
};

export async function listPartitions(
  path: string
): Promise<ListPartitionsResponse> {
  try {
    const apiUrl = process.env.API_URL || "http://localhost:8080";
    const response = await fetch(
      `${apiUrl}/disks/partitions?path=${encodeURIComponent(path)}`,
      {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
        cache: "no-store",
      }
    );

    if (!response.ok) {
      return {
        success: false,
        msg: `Error al obtener particiones: ${response.status} ${response.statusText}`,
      };
    }

    const data = await response.json();

    console.log(data);

    return {
      success: true,
      partitions: data.partitions || [],
    };
  } catch (error) {
    console.error("Error al obtener las particiones:", error);
    return {
      success: false,
      msg: `Error al obtener particiones: ${(error as Error).message}`,
    };
  }
}
